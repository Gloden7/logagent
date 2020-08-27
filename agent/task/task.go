package task

import (
	"context"
	"fmt"
	"logagent/util"
	"math/rand"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

type message map[string]interface{}

type closer interface {
	Close() error
}

type process func(msg message) (message, error)

// Task 任务结构体
type Task struct {
	samplingRate    float64
	goNum           int
	degradation     bool
	degradationChan chan bool
	logger          *zap.SugaredLogger
	msgs            chan message
	collector       func()
	processor       process
	handlers        []func(msg message)
	ctx             context.Context
	cancel          context.CancelFunc
	goCancels       []context.CancelFunc
	closers         []closer
}

// New 创建一个任务对象
func New(logger *zap.SugaredLogger, conf *Conf, degradation bool) *Task {
	ctx, cancel := context.WithCancel(context.Background())
	t := &Task{
		samplingRate: 1,
		logger:       logger,
		msgs:         make(chan message, 100),
		ctx:          ctx,
		cancel:       cancel,
		degradation:  degradation && !conf.NoDegradation,
	}
	t.initCollector(conf.Collector)

	if conf.GoNum > 1 && conf.GoNum < 10 {
		t.goNum = conf.GoNum
	}

	if len(conf.Parser.Mode) > 0 {
		t.initParser(conf.Parser)
	}
	if len(conf.Validators) > 0 {
		t.initGlobalValidators(conf.Validators)
	}
	if len(conf.Rewrites) > 0 {
		t.initRewriters(conf.Rewrites)
	}
	t.initHandlers(conf.Handlers)

	return t
}

// Run 任务运行
func (t *Task) Run() {
	defer func() {
		if err := recover(); err != nil {
			t.logger.Errorf("recovered:", err)
		}
	}()
	go t.collector()
	runFunc := t.getRunFunc()
	for i := 0; i < t.goNum; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		go runFunc(ctx)
		t.addGoCancel(cancel)
	}
	go t.watchMsgs(runFunc)
	if t.degradation {
		go t.genDegradationRate()
	}
	runFunc(t.ctx)
}

func (t *Task) getRunFunc() func(ctx context.Context) {
	if t.degradation && t.processor != nil {
		return func(ctx context.Context) {
			for {
				select {
				case msg := <-t.msgs:
					if <-t.degradationChan {
						continue
					}
					msg, err := t.processor(msg)
					if err != nil {
						t.logger.Error(err)
						if m, err := json.Marshal(msg); err == nil {
							t.logger.Error(util.Bytes2str(m))
						}
						continue
					}
					for _, handler := range t.handlers {
						handler(msg)
					}
				case <-ctx.Done():
					return
				}
			}
		}
	}
	if t.degradation && t.processor == nil {
		return func(ctx context.Context) {
			for {
				select {
				case msg := <-t.msgs:
					if <-t.degradationChan {
						continue
					}
					for _, handler := range t.handlers {
						handler(msg)
					}
				case <-ctx.Done():
					return
				}
			}
		}
	}
	if !t.degradation && t.processor != nil {
		return func(ctx context.Context) {
			for {
				select {
				case msg := <-t.msgs:
					msg, err := t.processor(msg)
					if err != nil {
						m, _ := json.Marshal(msg)
						t.logger.Errorf("%s, %s", m, err)
						continue
					}
					for _, handler := range t.handlers {
						handler(msg)
					}
				case <-ctx.Done():
					return
				}
			}
		}
	}
	return func(ctx context.Context) {
		for {
			select {
			case msg := <-t.msgs:
				for _, handler := range t.handlers {
					handler(msg)
				}
			case <-ctx.Done():
				return
			}
		}
	}
}

func (t *Task) getmsgsSize() int {
	return len(t.msgs)
}

func (t *Task) addGoCancel(c context.CancelFunc) {
	t.goCancels = append(t.goCancels, c)
}

func (t *Task) watchMsgs(runFunc func(ctx context.Context)) {
	for {
		dNum := t.getmsgsSize()/10 - t.goNum
		if dNum > 0 {
			for i := 0; i < dNum; i++ {
				ctx, cancel := context.WithCancel(context.Background())
				go runFunc(ctx)
				t.addGoCancel(cancel)
			}
		} else {
			for _, cancel := range t.goCancels[:-dNum] {
				cancel()
			}
			t.goCancels = t.goCancels[-dNum:]
		}
		t.goNum += dNum

		time.Sleep(3 * time.Second)
	}
}

func (t *Task) initParser(conf parserConf) {
	switch conf.Mode {
	case "csv":
		var delimiters string
		if len(conf.Delimiters) > 0 {
			delimiters = conf.Delimiters
		} else {
			delimiters = ","
		}
		colsLen := len(conf.Columns)
		if colsLen == 0 {
			t.configureFatal("csv parse", "columns")
		}

		t.processor = func(msg message) (message, error) {
			if message, ok := msg["message"].(string); ok {
				newMsg := strings.SplitN(message, delimiters, colsLen)
				for i, data := range newMsg {
					msg[conf.Columns[i]] = data
				}
			}
			return msg, nil
		}
	case "regex":
		if len(conf.Regex) == 0 {
			t.configureFatal("regex parse", "regex")
		}
		cmp, err := regexp.Compile(conf.Regex)
		if err != nil {
			t.logger.Fatal("invalid configuration `regex`")
		}

		t.processor = func(msg message) (message, error) {
			if message, ok := msg["message"].(string); ok {
				newMsg := cmp.FindStringSubmatch(message)
				groupNames := cmp.SubexpNames()
				for i, data := range newMsg {
					msg[groupNames[i]] = data
				}
			}
			return msg, nil
		}
	case "jsonify":
		t.processor = func(msg message) (message, error) {
			if message, ok := msg["message"].(string); ok {
				err := json.Unmarshal(util.Str2bytes(message), &msg)
				return msg, err
			}
			return msg, nil
		}
	default:
		t.logger.Fatalf("unsupported parser mode `%s`", conf.Mode)
	}
}

func (t *Task) initRewriters(rewritersConf []rewriterConf) {
	for _, conf := range rewritersConf {
		switch conf.Mode {
		case "set":
			if len(conf.Column) == 0 {
				t.configureFatal("set rewrite", "column")
			}
			t.setProcessor(func(msg message) (message, error) {
				msg[conf.Column] = conf.Value
				return msg, nil
			})
		case "subst":
			if len(conf.Column) == 0 {
				t.configureFatal("subst rewrite", "column")
			}
			if len(conf.Old) == 0 {
				t.configureFatal("subst rewrite", "old")
			}
			t.setProcessor(func(msg message) (message, error) {
				key, ok := msg[conf.Column].(string)
				if !ok {
					return msg, nil
				}
				msg[conf.Column] = strings.Replace(key, conf.Old, conf.Value, -1)
				return msg, nil
			})
		case "mapping":
			if len(conf.Column) == 0 {
				t.configureFatal("mapping rewrite", "column")
			}
			if conf.Mapping == nil {
				t.configureFatal("mapping rewrite", "mapping")
			}
			t.setProcessor(func(msg message) (message, error) {
				key, ok := msg[conf.Column].(string)
				if !ok {
					return msg, nil
				}
				value, ok := conf.Mapping[key]
				if !ok {
					return msg, fmt.Errorf("mapping rewriter column %s not found", key)
				}
				msg[conf.Column] = value
				return msg, nil
			})
		case "jsonify":
			if len(conf.Column) == 0 {
				t.configureFatal("jsonify rewrite", "column")
			}
			t.setProcessor(func(msg message) (message, error) {
				jsonStr, ok := msg[conf.Column].(string)
				if !ok {
					return msg, nil
				}
				tempMap := make(message)
				err := json.Unmarshal(util.Str2bytes(jsonStr), &tempMap)
				if err != nil {
					return msg, err
				}
				for k, v := range tempMap {
					msg[fmt.Sprintf("%s_%s", conf.Column, k)] = v
				}
				return msg, nil
			})
		case "command":
			if len(conf.Column) == 0 {
				t.configureFatal("command rewrite", "column")
			}
			if len(conf.Command) == 0 {
				t.configureFatal("subst rewrite", "command")
			}
			t.setProcessor(func(msg message) (message, error) {
				cmd := exec.Command(conf.Command)
				out, err := cmd.Output()
				if err != nil {
					return msg, err
				}
				msg[conf.Column] = util.Bytes2str(out)
				return msg, nil
			})
		case "splicing":
			if len(conf.Columns) == 0 {
				t.configureFatal("splicing rewrite", "column")
			}
			if len(conf.Key) == 0 {
				t.configureFatal("splicing rewrite", "key")
			}
			if len(conf.Delimiters) == 0 {
				conf.Delimiters = " "
			}
			t.setProcessor(func(msg message) (message, error) {
				values := []string{}
				for _, k := range conf.Columns {
					if v, ok := msg[k].(string); ok {
						values = append(values, v)
					}
				}
				msg[conf.Key] = strings.Join(values, conf.Delimiters)
				return msg, nil
			})
		default:
			t.logger.Fatalf("unsupported rewriter mode `%s`", conf.Mode)
		}
	}
}

func (t *Task) setProcessor(p process) {
	if &t.processor != nil {
		func(processor process) {
			t.processor = func(msg message) (message, error) {
				msg, err := processor(msg)
				if err != nil {
					return msg, err
				}
				return p(msg)
			}
		}(t.processor)
	} else {
		t.processor = p
	}
}

func (t *Task) addCloser(c closer) {
	t.closers = append(t.closers, c)
}

// Close 任务结束
func (t *Task) Close() error {
	for _, cancel := range t.goCancels {
		cancel()
	}
	t.cancel()
	time.Sleep(time.Millisecond * 500)
	for _, closer := range t.closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// SetSamplingRate 给任务设置采样率
func (t *Task) SetSamplingRate(s float64) {
	t.samplingRate = s
}

func (t *Task) genDegradationRate() {
	t.degradationChan = make(chan bool, 10)
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			t.degradationChan <- (rand.Float64() > t.samplingRate)
		}
	}
}

func (t *Task) configureFatal(model, confItem string) {
	t.logger.Fatalf("%s requires `%s` configuration.", model, confItem)
}
