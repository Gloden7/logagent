package task

import (
	"context"
	"encoding/json"
	"fmt"
	"logagent/util"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

type msg map[string]interface{}

type closer interface {
	Close() error
}

type process func(msg map[string]interface{}) map[string]interface{}

// Task 任务结构体
type Task struct {
	samplingRate    float64
	goNum           int
	degradation     bool
	logger          *zap.SugaredLogger
	msgs            chan msg
	degradationChan chan bool
	collector       func()
	processor       process
	handlers        []func(msg map[string]interface{})
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
		msgs:         make(chan msg, 100),
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
	if len(conf.Rewrites) > 0 {
		t.initRewriters(conf.Rewrites)
	}

	t.initHandlers(conf.Handlers)

	return t
}

// Run task run
func (t *Task) Run() {
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
	t.Close()
}

func (t *Task) getRunFunc() func(ctx context.Context) {
	if t.degradation {
		return func(ctx context.Context) {
			for {
				select {
				case msg := <-t.msgs:
					if <-t.degradationChan {
						continue
					}
					if t.processor != nil {
						msg = t.processor(msg)
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
				if t.processor != nil {
					msg = t.processor(msg)
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

func (t *Task) initCollector(conf collectorConf) {
	switch conf.Mode {
	case "api":
		t.setAPICollector(conf)
	case "syslog":
		t.setSyslogCollector(conf)
	default:
		t.logger.Panicf("Unsupported mode `%s`", conf.Mode)
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
			t.logger.Panic("Please configure columns")
		}

		t.processor = func(msg map[string]interface{}) map[string]interface{} {
			if message, ok := msg["message"].(string); ok {
				newMsg := strings.SplitN(message, delimiters, colsLen)
				for i, data := range newMsg {
					msg[conf.Columns[i]] = data
				}
			}
			return msg
		}
	case "regex":
		if len(conf.Regex) == 0 {
			t.logger.Panic("Please configure regex")
		}
		cmp, err := regexp.Compile(conf.Regex)
		if err != nil {
			t.logger.Panic("Please configure regex")
		}

		t.processor = func(msg map[string]interface{}) map[string]interface{} {
			if message, ok := msg["message"].(string); ok {
				newMsg := cmp.FindStringSubmatch(message)
				groupNames := cmp.SubexpNames()
				for i, data := range newMsg {
					msg[groupNames[i]] = data
				}
			}
			return msg
		}
	case "jsonify":
		t.processor = func(msg map[string]interface{}) map[string]interface{} {
			if message, ok := msg["message"].(string); ok {
				err := json.Unmarshal(util.Str2bytes(message), &msg)
				if err != nil {
					t.logger.Warn(err)
				}
			}
			return msg
		}
	default:
		t.logger.Panicf("Unsupported mode `%s`", conf.Mode)
	}
}

func (t *Task) initRewriters(rewritersConf []rewriterConf) {
	for _, conf := range rewritersConf {
		switch conf.Mode {
		case "set":
			if len(conf.Column) == 0 {
				return
			}
			t.setprocessor(func(msg map[string]interface{}) map[string]interface{} {
				msg[conf.Column] = conf.Value
				return msg
			})
		case "subst":
			if len(conf.Column) == 0 || len(conf.Old) == 0 {
				return
			}
			t.setprocessor(func(msg map[string]interface{}) map[string]interface{} {
				key, ok := msg[conf.Column].(string)
				if !ok {
					return msg
				}
				msg[conf.Column] = strings.Replace(key, conf.Old, conf.Value, -1)
				return msg
			})
		case "mapping":
			t.setprocessor(func(msg map[string]interface{}) map[string]interface{} {
				key, ok := msg[conf.Column].(string)
				if !ok {
					return msg
				}
				value, ok := conf.Mapping[key]
				if !ok {
					return msg
				}
				msg[conf.Column] = value
				return msg
			})
		case "jsonify":
			if len(conf.Column) == 0 {
				return
			}
			t.setprocessor(func(msg map[string]interface{}) map[string]interface{} {
				jsonStr, ok := msg[conf.Column].(string)
				if !ok {
					return msg
				}

				tempMap := make(map[string]interface{})
				err := json.Unmarshal(util.Str2bytes(jsonStr), &tempMap)
				if err != nil {
					t.logger.Warn(err)
					return msg
				}
				for k, v := range tempMap {
					msg[fmt.Sprintf("%s_%s", conf.Column, k)] = v
				}
				return msg
			})
		default:
			t.logger.Panicf("Unsupported mode `%s`", conf.Mode)
		}
	}
}

func (t *Task) setprocessor(p process) {
	if &t.processor != nil {
		func(processor process) {
			t.processor = func(msg map[string]interface{}) map[string]interface{} {
				msg = processor(msg)
				return p(msg)
			}
		}(t.processor)
	} else {
		t.processor = p
	}
}

func (t *Task) initHandlers(handlesConf []handlerConf) {
	for _, conf := range handlesConf {
		switch conf.Mode {
		case "stream":
			t.addStreamHandler(conf)
		case "file":
			t.addFileHandler(conf)
		case "database":
			t.addDBHandler(conf)
		default:
			t.logger.Panicf("Unsupported mode `%s`", conf.Mode)
		}
	}
}

func (t *Task) addHandle(handler func(msg map[string]interface{})) {
	t.handlers = append(t.handlers, handler)
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

	for _, closer := range t.closers {
		closer.Close()
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
