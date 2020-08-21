package task

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"logagent/tail"
	"net"
	"net/http"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (t *Task) setAPICollector(conf collectorConf) {
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
			data := message{
				"timestamp": time.Now(),
			}
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
		t.logger.Infof("api server start %s", conf.Addr)
		err := srv.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			t.logger.Panic(err)
		}
	}
}

func (t *Task) setSyslogCollector(conf collectorConf) {
	var end byte
	if len(conf.End) > 0 {
		end = []byte(conf.End)[0]
	} else {
		end = '\x00'
	}
	if len(conf.Addr) == 0 {
		conf.Addr = ":514"
	}

	if strings.Contains(conf.Protocol, "udp") {
		listener, err := net.ListenPacket("udp", conf.Addr)
		if err != nil {
			t.logger.Fatal(err)
		}
		defer listener.Close()

		t.collector = func() {
			t.logger.Infof("%s server start udp@%s", conf.Addr)
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
		return
	}

	listener, err := net.Listen("tcp", conf.Addr)
	if err != nil {
		t.logger.Fatal(err)
	}

	t.collector = func() {
		t.logger.Infof("syslog server start tcp@%s", conf.Addr)
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

				t.logger.Infof("client %s establishes connection", conn.RemoteAddr())
				go func() {
					defer conn.Close()
					reader := bufio.NewReader(conn)
					for {
						msg, err := decode(reader, end)
						if err == io.EOF {
							t.logger.Warn("client disconnect")
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
	}
}

func (t *Task) setFileCollector(conf collectorConf) {
	tails, err := tail.TailFile(conf.FileName, tail.Config{
		ReOpen:    !conf.NoReopen,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
	})

	if err != nil {
		t.logger.Fatalf("tailf err %s", err)
	}

	t.collector = func() {
		var ok bool
		var msg *tail.Line
		for {
			select {
			case <-t.ctx.Done():
				return
			case msg, ok = <-tails.Lines:
				if !ok {
					fmt.Printf("tail file close reopen, filename:%s\n", tails.Filename)
					time.Sleep(1000 * time.Millisecond)
					continue
				}
				t.msgs <- message{
					"message":   msg.Text,
					"timestamp": msg.Time,
				}
			}
		}
	}
}

func (t *Task) initCollector(conf collectorConf) {
	switch conf.Mode {
	case "api":
		t.setAPICollector(conf)
	case "syslog":
		t.setSyslogCollector(conf)
	case "file":
		t.setFileCollector(conf)
	default:
		t.logger.Fatalf("unsupported collector mode `%s`", conf.Mode)
	}
}
