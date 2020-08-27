package task

import (
	"context"
	"fmt"
	"logagent/kafka"
	"os"
	"strings"
	"time"

	"github.com/Shopify/sarama"

	"github.com/natefinch/lumberjack"
)

type handler func(msg message)

func (t *Task) newFileHandler(conf handlerConf) handler {
	if len(conf.FileName) == 0 {
		t.configureFatal("file handle", "filename")
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

	templateFunc := getTemplateFunc(conf.Template)

	return func(msg message) {
		text := templateFunc(msg)
		_, err := fmt.Fprintln(f, text)
		if err != nil {
			t.logger.Error(err)
		}
	}
}

func (t *Task) newDBHandler(conf handlerConf) handler {
	if len(conf.Table) == 0 {
		t.configureFatal("database handle", "table")
	}
	if len(conf.Columns) == 0 {
		t.configureFatal("database handle", "columns")
	}
	if len(conf.Fields) != 0 {
		if len(conf.Fields) != len(conf.Columns) {
			t.logger.Fatal("invalid `fields` configuration")
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
		t.logger.Fatal("bad database URI")
	}
	dn := temp[0]
	dsn := strings.TrimLeft(temp[1], "/")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	switch dn {
	case "postgresql":
		dn = "pg"
		dsn = conf.URI
	case "mysql":
		i := strings.IndexByte(dsn, '@') + 1
		j := strings.IndexByte(dsn, '/')
		if i == 0 || j == -1 {
			t.logger.Fatal("bad database URI")
		}
		dsn = dsn[:i] + "tcp(" + dsn[i:j] + ")" + dsn[j:]
	}

	dbInit := make(chan bool)
	go func() {
		select {
		case <-ctx.Done():
			t.logger.Fatal("database connection timed out")
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
			t.logger.Fatal(err)
		}
	}
	sql := genInsertSQL(dn, conf.Table, conf.Columns)
	stmt, err := database.Prepare(sql)
	if err != nil {
		t.logger.Fatal(err)
	}
	t.addCloser(stmt)

	sortFunc := genSortFunc(conf.Columns)

	return func(msg message) {
		// fmt.Println(msg["timestamp"])
		insertData := sortFunc(msg)
		fmt.Println(insertData)
		_, err = stmt.Exec(insertData...)
		if err != nil {
			t.logger.Error(err)
		}
	}
}

func (t *Task) newStreamHandler(conf handlerConf) handler {
	templateFunc := getTemplateFunc(conf.Template)
	return func(msg message) {
		_, err := fmt.Fprintln(os.Stdout, templateFunc(msg))
		if err != nil {
			t.logger.Error(err)
		}
	}
}

func (t *Task) newKafkaHandle(conf handlerConf) handler {
	if len(conf.Addrs) == 0 {
		t.configureFatal("kafka handler", "addrs")
	}

	var timeout time.Duration
	if conf.Timeout > 0 {
		timeout = time.Second * time.Duration(conf.Timeout)
	} else {
		timeout = time.Second * 10
	}

	p, err := kafka.Producer(conf.RequiredAcks, conf.Addrs, timeout)

	if err != nil {
		t.logger.Fatal(err)
	}

	var topic string
	if len(conf.Topic) > 0 {
		topic = conf.Topic
	} else {
		topic = "log_agent"
	}

	templateFunc := getTemplateFunc(conf.Template)

	return func(msg message) {
		text := templateFunc(msg)
		kafkaMsg := &sarama.ProducerMessage{}
		kafkaMsg.Topic = topic
		kafkaMsg.Value = sarama.ByteEncoder(text)
		_, _, err := p.SendMessage(kafkaMsg)
		if err != nil {
			t.logger.Error(err)
		}
	}
}

func (t *Task) addHandle(handler func(msg message)) {
	t.handlers = append(t.handlers, handler)
}

func (t *Task) initHandler(conf handlerConf) handler {
	switch conf.Mode {
	case "stream":
		return t.newStreamHandler(conf)
	case "file":
		return t.newFileHandler(conf)
	case "database":
		return t.newDBHandler(conf)
	case "kafka":
		return t.newKafkaHandle(conf)
	default:
		t.logger.Fatalf("unsupported handle mode `%s`", conf.Mode)
		return nil
	}
}

func (t *Task) initHandlers(handlesConf []handlerConf) {
	for _, conf := range handlesConf {
		vs := t.initValidators(conf.Validators)
		h := t.initHandler(conf)
		if len(vs) > 0 {
			t.addHandle(func(msg message) {
				for _, v := range vs {
					if err := v(msg); err != nil {
						return
					}
				}
				h(msg)
			})
		} else {
			t.addHandle(h)
		}
	}
}
