package lnksworks

import (
	"net/http"
	"strings"
	"time"
)

//Server conveniance struct wrapping arround *http.Server
type Server struct {
	port       string
	svr        *http.Server
	istls      bool
	certFile   string
	keyFile    string
	serveQueue chan func()
}

//NewServer return *Server
//istls is tls listener
//certfile string - path to certificate file
//keyFile string - path to key file
func NewServer(port string, istls bool, certFile string, keyFile string) *Server {
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	var svr = &Server{port: port, istls: istls, certFile: certFile, keyFile: keyFile, serveQueue: make(chan func())}
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

//Listen start listening for connection(s)
func (svr *Server) Listen() (err error) {
	svr.svr = &http.Server{Addr: svr.port, Handler: svr, ReadHeaderTimeout: 3 * 1024 * time.Millisecond, ReadTimeout: 30 * 1024 * time.Millisecond, WriteTimeout: 60 * 1024 * time.Millisecond}
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

//ServeHTTP server http.Handler interface implementation of ServeHTTP
func (svr *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	done := make(chan bool, 1)
	defer close(done)
	svr.serveQueue <- func() {
		routes.serve(svr, w, r)
		done <- true
	}
	<-done
}
