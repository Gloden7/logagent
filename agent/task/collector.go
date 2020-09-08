package task

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"logagent/kafka"
	"logagent/tail"
	"logagent/util"
	"net"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	jsoniter "github.com/json-iterator/go"
	"github.com/radovskyb/watcher"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (t *Task) setAPICollector(conf collectorConf) {
	if len(conf.Addr) == 0 {
		t.configureFatal("api collector", "addr")
	}
	var url string
	if len(conf.URL) > 0 {
		if !strings.HasPrefix(conf.URL, "/") {
			conf.URL = "/" + conf.URL
		}
		url = conf.URL
	} else {
		url = "/logs"
	}
	srv := getSrv(conf.Addr)

	yes, _ := json.Marshal(map[string]interface{}{
		"code":    0,
		"message": "successfully",
	})
	no, _ := json.Marshal(map[string]interface{}{
		"code":    1,
		"message": "bad request format",
	})

	srv.addHandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			data := message{
				"timestamp": time.Now(),
			}
			decode := json.NewDecoder(r.Body)
			err := decode.Decode(&data)

			w.Header().Set("Content-Type", "application/json")
			if err != nil {
				t.logger.Warn(err)
				w.WriteHeader(400)
				w.Write(no)
				return
			}
			data["device_id"] = t.deviceID
			t.msgs <- data
			w.WriteHeader(201)
			w.Write(yes)
		}
	})

	t.addCloser(srv.server)

	t.collector = func() {
		if err := srv.start(); err != nil && err != http.ErrServerClosed {
			t.logger.Fatal(err)
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

		t.collector = func() {
			defer listener.Close()
			t.logger.Infof("syslog server start udp@%s", conf.Addr)
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
					msg["device_id"] = t.deviceID
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
		defer listener.Close()
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
	if len(conf.Filename) == 0 {
		t.configureFatal("file Collector", "filename")
	}

	tailf, err := tail.TailFile(conf.Filename, tail.Config{
		Follow:    true,
		ReOpen:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: true,
	})
	if err != nil {
		t.logger.Fatal(err)
	}

	t.collector = func() {
		var ok bool
		var msg *tail.Line

		for {
			select {
			case <-t.ctx.Done():
				return
			case msg, ok = <-tailf.Lines:
				if !ok {
					fmt.Printf("tail file close reopen, filename:%s\n", tailf.Filename)
					time.Sleep(1000 * time.Millisecond)
					continue
				}
				t.msgs <- message{
					"message":   msg.Text,
					"timestamp": msg.Time,
					"device_id": t.deviceID,
				}
			}
		}
	}
}

func (t *Task) setKafkaCollector(conf collectorConf) {
	if len(conf.Addrs) == 0 {
		t.configureFatal("kafka collector", "addrs")
	}
	var timeout time.Duration
	if conf.Timeout > 0 {
		timeout = time.Second * time.Duration(conf.Timeout)
	} else {
		timeout = time.Second * 10
	}

	c, err := kafka.Consumer(conf.Addrs, timeout)

	if err != nil {
		t.logger.Fatal(err)
	}

	pl, err := c.Partitions(conf.Topic)
	if err != nil {
		t.logger.Fatal(err)
	}

	var offset int64
	if offsetContent, err := ioutil.ReadFile("./offset"); err == nil {
		off, err := strconv.Atoi(util.Bytes2str(offsetContent))
		if err != nil {
			t.logger.Fatal(err)
		}
		offset = int64(off)
	} else {
		offset = sarama.OffsetNewest
	}

	t.collector = func() {
		// defer c.Close()
		for _, partition := range pl {
			//ConsumePartition方法根据主题，分区和给定的偏移量创建创建了相应的分区消费者
			//如果该分区消费者已经消费了该信息将会返回error
			//sarama.OffsetNewest:表明了为最新消息
			pc, err := c.ConsumePartition(conf.Topic, int32(partition), offset)
			if err != nil {
				t.logger.Fatal(err)
			}
			// t.addCloser(pc)

			kafkaMsg := pc.Messages()
			var msg *sarama.ConsumerMessage
			for {
				select {
				case <-t.ctx.Done():
					if msg != nil {
						ioutil.WriteFile("./offset", []byte(fmt.Sprintf("%d", msg.Offset+1)), 0666)
					}
					return
				case msg = <-kafkaMsg:
					//Messages()该方法返回一个消费消息类型的只读通道，由代理产生
					t.msgs <- message{
						"message":   util.Bytes2str(msg.Value),
						"timestamp": time.Now(),
						"device_id": t.deviceID,
					}
				}
			}
		}
	}
}

func (t *Task) setDirCollector(conf collectorConf) {
	if len(conf.Filename) == 0 {
		t.configureFatal("dir collector", "filename")
	}

	dir, filename := filepath.Split(conf.Filename)
	r := regexp.MustCompile(fmt.Sprintf("^%s$", filename))

	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Create)
	filter := watcher.RegexFilterHook(r, false)
	w.AddFilterHook(filter)

	if err := w.Add(dir); err != nil {
		t.logger.Fatal(err)
	}

	t.collector = func() {
		go func() {
			for {
				select {
				case e := <-w.Event:
					if !e.IsDir() {
						b, err := ioutil.ReadFile(e.Path)
						if err != nil {
							t.logger.Error(err)
							continue
						}
						t.msgs <- message{
							"message":   util.Bytes2str(b),
							"timestamp": time.Now(),
							"device_id": t.deviceID,
						}
					}
				case err := <-w.Error:
					t.logger.Error(err)
				case <-w.Closed:
					return
				case <-t.ctx.Done():
					w.Close()
				}
			}
		}()

		if err := w.Start(time.Millisecond * 100); err != nil {
			t.logger.Fatal(err)
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
	case "kafka":
		t.setKafkaCollector(conf)
	case "dir":
		t.setDirCollector(conf)
	default:
		t.logger.Fatalf("unsupported collector mode `%s`", conf.Mode)
	}
}
