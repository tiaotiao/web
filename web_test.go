package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

var testStr string

type middWareAuth struct{}

func (md *middWareAuth) Name() string {
	return "auth"
}

func (md *middWareAuth) ServeMiddleware(c *Context) error {
	testStr += "auth "
	return nil
}

// func (md *middWareAuth) AfterMiddleware(w http.ResponseWriter, r *http.Request) {}

type middWareLog struct {
}

func (md *middWareLog) Name() string {
	return "log"
}

func (md *middWareLog) ServeMiddleware(c *Context) error {
	testStr += "log "
	return nil
}

///////////////////////////////////////////////////////////////////////

type testRouterHandle struct{}

func (h *testRouterHandle) Post(c *Context) interface{} {
	testStr += "handlerPost "
	return map[string]interface{}{"msg": "ok"}
}

func (h *testRouterHandle) Get(c *Context) interface{} {
	testStr += "handlerGet "
	return "ok"
}

func (h *testRouterHandle) Put(c *Context) interface{} {
	testStr += "handlerPut "
	return nil
}

func (h *testRouterHandle) Delete(c *Context) interface{} {
	testStr += "handlerDelete "
	return "ok"
}

func justR(method, url string) *http.Request {
	r, _ := http.NewRequest(method, url, nil)
	return r
}

func TestWeb(t *testing.T) {
	testCases := []struct {
		r          *http.Request
		statusCode int
		content    string
	}{
		{
			justR("POST", "http://localhost:8095/routerPost"),
			http.StatusOK,
			`{"msg":"ok"}`,
		},
		{
			justR("GET", "http://localhost:8095/routerGet"),
			http.StatusOK,
			"ok",
		},
		{
			justR("PUT", "http://localhost:8095/routerPut"),
			http.StatusOK,
			"",
		},
		{
			justR("DELETE", "http://localhost:8095/routerDelete"),
			http.StatusOK,
			"ok",
		},
		// {
		// 	justR("PATCH", "http://localhost:8095/routerDelete"),
		// 	http.StatusMethodNotAllowed,
		// 	`{"result":"method not allowed","msg":"method not allowed"}`,
		// },
	}

	r := NewWeb()
	r.Append(new(middWareAuth))
	// r.Append(new(middWareLog))

	h := new(testRouterHandle)
	client := &http.Client{}
	r.Handle("GET", "/routerGet", h.Get)
	r.Handle("PUT", "/routerPut", h.Put)
	r.Handle("POST", "/routerPost", h.Post)
	r.Handle("DELETE", "/routerDelete", h.Delete)

	go http.ListenAndServe(":8095", r)

	for _, tc := range testCases {
		resp, err := client.Do(tc.r)
		checkErr(err)
		checkResponse(resp, tc.statusCode, tc.content, t)
	}

	// sub router
	sr := r.SubRouter("sub")
	m2 := sr.Handle("GET", "/routerHandle", h.Get)
	m2.Append(new(middWareLog))
	// m2.Remove("auth")
	resp, err := http.Get("http://localhost:8095/sub/routerHandle")
	checkErr(err)
	checkResponse(resp, http.StatusOK, "ok", t)

	// not found
	resp, err = http.Get("http://localhost:8095/0")
	checkErr(err)
	checkResponse(resp, http.StatusNotFound, "404 page not found\n", t)

	// check middleware
	// if wantStr := "auth log handlerPost auth log handlerGet auth log handlerPut auth log handlerDelete log handlerGet "; wantStr != testStr {
	// 	t.Errorf("testStr = '%s'; want '%s'", testStr, wantStr)
	// }

	// prefix router
	r.SubRouter("pathPerfix")
	r.Handle("GET", "/prefixrouter/*", h.Get) // Handle a path end with '/*'
	resp, err = http.Get("http://localhost:8095/prefixrouter/123123/asdasdf/123123/sdgf")
	checkErr(err)
	checkResponse(resp, http.StatusOK, "ok", t)
	resp, err = http.Get("http://localhost:8095/prefixrouter/123123/asdasdf")
	checkErr(err)
	checkResponse(resp, http.StatusOK, "ok", t)
	resp, err = http.Get("http://localhost:8095/prefixrouter") // Path with out trailing '/' will not be found.
	checkErr(err)
	checkResponse(resp, http.StatusNotFound, "404 page not found\n", t)
}

// helper
func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func checkResponse(resp *http.Response, statusCode int, expctContent string, t *testing.T) {
	str, err := ioutil.ReadAll(resp.Body)
	checkErr(err)
	if resp.StatusCode != statusCode {
		t.Errorf("response statusCode = %d; want %d; url=%s", resp.StatusCode, statusCode, resp.Request.URL.String())
	}
	if expctContent != string(str) {
		t.Errorf("respone body = %s; want %s; url=%s", str, expctContent, resp.Request.URL.String())
	}
}
