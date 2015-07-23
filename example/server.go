package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tiaotiao/web"
)

type Server struct {
	addr string
	web  *web.Web
	api  *Api
}

func NewServer(addr string) *Server {
	s := new(Server)
	s.addr = addr
	s.web = web.NewWeb()
	s.api = NewApi()
	return s
}

func (s *Server) Run() error {
	s.init()

	println("server start", s.addr)

	err := s.web.ListenAndServe("tcp", s.addr)

	println("server stoped")

	return err
}

func (s *Server) init() {
	s.web.SetLogger(s)
	s.registerURLs()
}

func (s *Server) registerURLs() {
	s.web.HandleFunc("GET", "/api/message", s.api.GetMessage)
	s.web.HandleFunc("GET", "/api/message/list", s.api.GetMessages)
	s.web.HandleFunc("GET", "/api/message/add", s.api.PostMessage)
}

func (s *Server) OnLog(r *http.Request, start time.Time, used time.Duration, code int, result interface{}) {
	t := start.Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %d - %s %s - %s - %v\n", t, code, r.Method, r.URL.Path, used, result)
}
