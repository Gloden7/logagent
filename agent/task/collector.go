package task

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	pb "logagent/agent/task/plugins"

	"github.com/hpcloud/tail"
	"google.golang.org/grpc"
)

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
			data := map[string]interface{}{
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
		t.logger.Infof("API server start %s", conf.Addr)
		err := srv.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			t.logger.Panicf("API server start err %s", err)
		}
	}
}

func (t *Task) setGRPCCollector(conf collectorConf) {
	listener, err := net.Listen("tcp", conf.Addr)
	if err != nil {
		t.logger.Panic(err)
	}
	var opts []grpc.ServerOption
	s := grpc.NewServer(opts...)
	pb.RegisterLoggerServer(s, &server{
		t: t,
	})
	t.collector = func() {
		if err := s.Serve(listener); err != nil {
			t.logger.Panicf("failed to serve: %v", err)
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
			t.logger.Panic(err)
		}
		defer listener.Close()
		t.logger.Infof("%s server start %s", conf.Protocol, conf.Addr)

		t.collector = func() {
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
		t.logger.Panic(err)
	}
	t.logger.Infof("%s server start %s", conf.Protocol, conf.Addr)

	t.collector = func() {
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
	}
}

func (t *Task) setFileCollector(conf collectorConf) {
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
					"message":   line.Text,
					"timestamp": line.Time,
				}
			}
		}
	}
}
