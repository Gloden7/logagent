package task

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"logagent/util"
	"math/rand"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	pb "logagent/agent/task/plugins"

	"github.com/hpcloud/tail"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type msg map[string]interface{}

type closer interface {
	Close() error
}

type process func(msg map[string]interface{}) map[string]interface{}

// Task log task
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

// New create log task
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

	if conf.GoNum > 1 {
		if conf.GoNum < 10 {
			t.goNum = conf.GoNum - 1
		} else {
			t.goNum = 9
		}
	} else {
		t.goNum = 0
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
	for _, closer := range t.closers {
		closer.Close()
	}
	close(t.msgs)
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
		var url string
		if len(conf.URL) > 0 {
			if !strings.HasPrefix(conf.URL, "/") {
				conf.URL = "/" + conf.URL
			}
			url = conf.URL
		} else {
			url = "/logs"
		}

		serverMux := http.NewServeMux()

		serverMux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				var data map[string]interface{}
				decode := json.NewDecoder(r.Body)
				err := decode.Decode(&data)

				if err != nil {
					t.logger.Warn(err)
					w.WriteHeader(400)
					return
				}

				t.msgs <- data
				w.WriteHeader(201)
			}
		})

		srv := http.Server{
			Addr:    conf.Addr,
			Handler: serverMux,
		}

		t.addCloser(&srv)

		t.collector = func() {
			t.logger.Infof("API server start %s", conf.Addr)
			err := srv.ListenAndServe()

			if err != nil && err != http.ErrServerClosed {
				t.logger.Panicf("API server start err %s", err)
			}
		}
	case "file":
		tails, err := tail.TailFile(conf.FileName, tail.Config{
			ReOpen:    !conf.NoReopen,
			Follow:    true,
			Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
			MustExist: false,
			Poll:      true,
		})

		if err != nil {
			t.logger.Panicf("tailf err %s", err)
		}

		t.collector = func() {
			var ok bool
			var line *tail.Line

			for {
				select {
				case <-t.ctx.Done():
					return
				case line, ok = <-tails.Lines:
					if !ok {
						fmt.Printf("tail file close reopen, filename: %s\n", tails.Filename)
						time.Sleep(1000 * time.Millisecond)
						continue
					}
					t.msgs <- map[string]interface{}{
						"message": line.Text,
						"time":    line.Time,
					}
				}
			}
		}

	case "syslog":
		var end byte
		if len(conf.End) > 0 {
			end = []byte(conf.End)[0]
		} else {
			end = '\x00'
		}
		if len(conf.Addr) == 0 {
			conf.Addr = ":514"
		}

		t.collector = func() {
			if !strings.Contains(conf.Protocol, "udp") {
				listener, err := net.Listen("tcp", conf.Addr)
				if err != nil {
					t.logger.Panic(err)
				}

				t.logger.Infof("%s server start %s", conf.Protocol, conf.Addr)

				for {
					select {
					case <-t.ctx.Done():
						return
					default:
						conn, err := listener.Accept() // 建立连接
						if err != nil {
							t.logger.Warn(err)
							continue
						}
						defer conn.Close()

						t.logger.Infof("Client %s establishes connection", conn.RemoteAddr())
						go func() {
							defer conn.Close()
							reader := bufio.NewReader(conn)
							for {
								msg, err := decode(reader, end)
								if err == io.EOF {
									t.logger.Warn("Client disconnect")
									break
								}
								if err != nil {
									t.logger.Warn(err)
									continue
								}
								t.msgs <- msg
							}
						}()
					}
				}
			} else {
				listener, err := net.ListenPacket("udp", conf.Addr)
				if err != nil {
					t.logger.Panic(err)
				}
				defer listener.Close()
				t.logger.Infof("%s server start %s", conf.Protocol, conf.Addr)

				for {
					select {
					case <-t.ctx.Done():
						return
					default:
						var buf [1024]byte
						n, _, err := listener.ReadFrom(buf[:])
						if err != nil {
							t.logger.Warn(err)
							continue
						}
						reader := bufio.NewReader(bytes.NewReader(buf[:n]))
						msg, err := decode(reader, end)
						if err != nil {
							t.logger.Warn(err)
							continue
						}
						t.msgs <- msg
					}
				}
			}
		}
	case "grpc":
		t.collector = func() {
			listener, err := net.Listen("tcp", conf.Addr)
			if err != nil {
				t.logger.Panic(err)
			}
			var opts []grpc.ServerOption
			s := grpc.NewServer(opts...)
			pb.RegisterLoggerServer(s, &server{
				t: t,
			})
			if err := s.Serve(listener); err != nil {
				t.logger.Panicf("failed to serve: %v", err)
			}
		}
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
			var template []byte
			if len(conf.Template) > 0 {
				template = util.Str2bytes(conf.Template)
			} else {
				template = util.Str2bytes("${MESSAGE}")
			}
			t.addHandle(func(msg map[string]interface{}) {
				content := templates(template, msg)
				_, err := os.Stdout.Write(content)
				if err != nil {
					t.logger.Warn(err)
				}
			})
		case "file":
			if len(conf.FileName) == 0 {
				t.logger.Panic("File configuration `path` cannot be empty")
			}

			var maxSize int
			if conf.MaxSize > 0 {
				maxSize = conf.MaxSize
			} else {
				maxSize = 10
			}

			f := &lumberjack.Logger{
				Filename:   conf.FileName,
				MaxSize:    maxSize,
				MaxBackups: conf.MaxBackups,
				MaxAge:     conf.MaxAge,
				Compress:   conf.Compress,
			}
			t.addCloser(f)

			var template []byte
			if len(conf.Template) > 0 {
				template = util.Str2bytes(conf.Template)
			} else {
				template = util.Str2bytes("${MESSAGE}")
			}

			t.addHandle(func(msg map[string]interface{}) {
				content := templates(template, msg)
				_, err := f.Write(content)
				if err != nil {
					t.logger.Warn(err)
				}
			})
		case "database":
			if len(conf.Table) == 0 {
				t.logger.Panic("Database configuration `table` cannot be empty")
			}

			if len(conf.Columns) == 0 {
				t.logger.Panic("Database configuration `columns` cannot be empty")
			}
			if len(conf.Fields) != 0 {
				if len(conf.Fields) != len(conf.Columns) {
					t.logger.Panic("Columns fields have different lengths")
				}
			}

			var timeout time.Duration
			if conf.Timeout > 0 {
				timeout = time.Second * time.Duration(conf.Timeout)
			} else {
				timeout = time.Second * 10
			}

			temp := strings.SplitN(conf.URI, ":", 2)
			if len(temp) < 2 {
				t.logger.Panic("Bad database URI")
			}
			dn := temp[0]
			dsn := strings.TrimLeft(temp[1], "/")

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			if dn == "mongodb" {
				i := strings.LastIndexByte(conf.URI, '/')
				if i == -1 {
					t.logger.Panic("Bad database URI")
				}
				dsn = conf.URI[:i]
				database := conf.URI[i+1:]

				clt, err := initMongoDB(ctx, dsn)
				if err != nil {
					t.logger.Panicf("Mongodb connection failed %s", err)
				}
				t.addCloser(clt)

				collection := clt.Client.Database(database).Collection(conf.Table)
				sortFunc := genSortFunc(conf.Columns)

				t.addHandle(func(data map[string]interface{}) {
					insertData := sortFunc(data)
					_, err = collection.InsertOne(context.TODO(), insertData)
					if err != nil {
						t.logger.Warn(err)
					}
				})

				return
			}

			switch dn {
			case "postgresql":
				dn = "pg"
				dsn = conf.URI
			case "mysql":
				i := strings.IndexByte(dsn, '@') + 1
				j := strings.IndexByte(dsn, '/')
				if i == 0 || j == -1 {
					t.logger.Panic("Bad database URI")
				}
				dsn = dsn[:i] + "tcp(" + dsn[i:j] + ")" + dsn[j:]
			}

			dbInit := make(chan bool)
			go func() {
				select {
				case <-ctx.Done():
					t.logger.Panic("Database connection timed out")
				case <-dbInit:
					break
				}
			}()

			database, err := initDatabase(dn, dsn)
			dbInit <- true
			if err != nil {
				t.logger.Panic(err)
			}
			t.addCloser(database)

			if len(conf.Fields) > 0 {
				err = createTable(database, dn, conf.Table, conf.Fields)
				if err != nil {
					t.logger.Panic(err)
				}
			}
			sql := genInsertSQL(dn, conf.Table, conf.Columns)
			stmt, err := database.Prepare(sql)
			if err != nil {
				t.logger.Panic(err)
			}
			t.addCloser(stmt)

			sortFunc := genSortFunc(conf.Columns)

			t.addHandle(func(data map[string]interface{}) {
				insertData := sortFunc(data)
				_, err = stmt.Exec(insertData...)
				if err != nil {
					t.logger.Warn(err)
				}
			})

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

// Close task exit
func (t *Task) Close() error {
	for _, cancel := range t.goCancels {
		cancel()
	}
	t.cancel()
	return nil
}

// SetSamplingRate setSamplingRate
func (t *Task) SetSamplingRate(s float64) {
	t.samplingRate = s
}

func (t *Task) genDegradationRate() {
	t.degradationChan = make(chan bool, 10)
	for {
		t.degradationChan <- (rand.Float64() > t.samplingRate)
	}
}
