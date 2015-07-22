package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var testALStr string

type testSHMiddleware struct{}

func (m *testSHMiddleware) Name() string {
	return "testSH"
}

func (m *testSHMiddleware) ServeMiddleware(c *Context) error {
	testALStr += "testSH"
	return nil
}

type TestHandler struct{}

func (th TestHandler) Get(c *Context) interface{} {
	testALStr += "testSH"
	return struct {
		Name string `json:"Name"`
		Tel  int    `json:"Tel"`
	}{"Tester", 123123}
}

type testLogger struct{}

func (ta testLogger) OnLog(r *http.Request, start time.Time, used time.Duration, code int) {
	testALStr += fmt.Sprintf("%s-%v-%v-%v", r.URL, start, used, code)
}

func TestServeHttp(t *testing.T) {
	th := TestHandler{}

	mm := newMiddlewaresManager()
	m := testSHMiddleware{}
	mm.Append(&m)

	tal := testLogger{}

	wh := NewWebHandler(th.Get, mm, NewDefaultResponser(), tal)
	rw := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://local/test", strings.NewReader("t1=v1"))
	checkErr(err)
	wh.ServeHTTP(rw, r)
	if !strings.Contains(testALStr, "testSH") {
		t.Errorf("testALStr=%s; want contains %s", testALStr, "testSH")
	}
	rspD, _ := ioutil.ReadAll(rw.Body)
	if wantD := `{"Name":"Tester","Tel":123123}`; string(rspD) != wantD {
		t.Errorf("responseBody=%s; want %s", rspD, wantD)
	}
}
