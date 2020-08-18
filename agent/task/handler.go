package task

import (
	"context"
	"fmt"
	"logagent/util"
	"os"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
)

func (t *Task) addFileHandler(conf handlerConf) {
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

	if len(conf.Template) == 0 {
		conf.Template = "${MESSAGE}"
	}
	templateFunc := getTemplateFunc(conf.Template)

	t.addHandle(func(msg map[string]interface{}) {
		text := templateFunc(msg)
		_, err := f.Write(util.Str2bytes(text))
		if err != nil {
			t.logger.Warn(err)
		}
	})
}

func (t *Task) addDBHandler(conf handlerConf) {
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

	t.addHandle(func(msg map[string]interface{}) {
		insertData := sortFunc(msg)
		_, err = stmt.Exec(insertData...)
		if err != nil {
			t.logger.Warn(err)
		}
	})
}

func (t *Task) addStreamHandler(conf handlerConf) {
	if len(conf.Template) == 0 {
		conf.Template = "${MESSAGE}"
	}

	templateFunc := getTemplateFunc(conf.Template)
	t.addHandle(func(msg map[string]interface{}) {
		_, err := fmt.Fprint(os.Stdout, templateFunc(msg))
		if err != nil {
			t.logger.Warn(err)
		}
	})
}
