package task

import (
	"net/http"
	"sync"
)

var srvs = sync.Map{}

type srv struct {
	running  bool
	server   *http.Server
	serveMux *http.ServeMux
}

func (s *srv) start() error {
	if s.running {
		return nil
	}
	s.running = true
	return s.server.ListenAndServe()
}

func (s *srv) addHandleFunc(url string, f func(w http.ResponseWriter, r *http.Request)) {
	s.serveMux.HandleFunc(url, f)
}

func getSrv(addr string) *srv {
	if s, ok := srvs.Load(addr); ok {
		return s.(*srv)
	}
	serveMux := http.NewServeMux()
	server := &http.Server{
		Addr:    addr,
		Handler: serveMux,
	}
	s := &srv{
		server:   server,
		serveMux: serveMux,
	}
	srvs.Store(addr, s)
	return s
}
