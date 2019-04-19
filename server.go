package lnksworks

import (
	"net/http"
	"strings"
	"time"
)

type Server struct {
	port       string
	svr        *http.Server
	istls      bool
	certFile   string
	keyFile    string
	serveQueue chan func()
}

func NewServer(port string, istls bool, certFile string, keyFile string) *Server {
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	var svr *Server = &Server{port: port, istls: istls, certFile: certFile, keyFile: keyFile, serveQueue: make(chan func())}

	return svr
}

func (svr *Server) processServes() {
	for {
		select {
		case serve := <-svr.serveQueue:
			go serve()
		}
	}
}

func (svr *Server) Listen() (err error) {
	svr.svr = &http.Server{Addr: svr.port, Handler: svr, ReadHeaderTimeout: 3 * 1024 * time.Millisecond, ReadTimeout: 3 * 1024 * time.Millisecond, WriteTimeout: 1024 * time.Millisecond}
	if svr.istls {
		go svr.processServes()
		err = svr.svr.ListenAndServeTLS(svr.certFile, svr.keyFile)
	} else {
		go svr.processServes()
		err = svr.svr.ListenAndServe()
	}
	close(svr.serveQueue)
	return err
}

func (svr *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	done := make(chan bool, 1)
	defer close(done)
	svr.serveQueue <- func() {
		routes.serve(svr, w, r)
		done <- true
	}
	<-done
}
