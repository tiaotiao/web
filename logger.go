package web

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Logger interface {
	OnLog(r *http.Request, start time.Time, used time.Duration, code int, result interface{})
}

type StdLogger struct {
}

func NewStdLogger() *StdLogger {
	return new(StdLogger)
}

func (l *StdLogger) OnLog(r *http.Request, start time.Time, used time.Duration, code int, result interface{}) {
	var post string
	if r.Method == "POST" {
		for k, v := range r.PostForm {
			if post != "" {
				post += "&"
			}
			post += fmt.Sprintf("%s=%s", k, strings.Join(v, ","))
		}
		if post != "" {
			post = " " + post
		}
	}
	fmt.Printf("[%s] %3d - %4s %s%s - %dns %v\n", start.Format("0102 15:04:05"), code, r.Method, r.RequestURI, post, used, result)
}
