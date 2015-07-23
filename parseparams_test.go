package web

import (
	"bytes"
	"github.com/gorilla/mux"
	"net/http"
	"testing"
)

var context *Context

func serve(w http.ResponseWriter, r *http.Request) {

	context, _ = NewContext(w, r)

	ParseParams(context)
}

func checkparam(key string, val string, t *testing.T) {
	if v, ok := context.GetString(key); ok {
		if v != val {
			t.Error("param '", key, "' !=", val)
		}
	} else {
		t.Error("param '", key, "' not found")
	}
}

func TestParseParamsMiddleware(t *testing.T) {

	router := mux.NewRouter()
	router.HandleFunc("/user/{uid:[0-9]+}/something", serve)
	router.HandleFunc("/postfile", serve)

	// request
	body := bytes.NewBuffer([]byte("body1=1000&body2=a,b,c&body3=a%3d1%26b%3d2"))
	req, err := http.NewRequest("POST", "http://qing.wps.cn/user/8533219/something?name=tom&ok=true&count=10&query=q%3dgolang%2burlencode%26gws_rd%3dssl", body)
	if err != nil {
		t.Fatal(err.Error())
	}

	router.ServeHTTP(nil, req)

	// check
	checkparam("name", "tom", t)
	checkparam("ok", "true", t)
	checkparam("count", "10", t)
	checkparam("uid", "8533219", t)
	checkparam("body1", "1000", t)
	checkparam("body2", "a,b,c", t)
	checkparam("body3", "a=1&b=2", t)
	checkparam("query", "q=golang+urlencode&gws_rd=ssl", t)

	// multipart request
	postdata := []byte(`--foo
Content-Disposition: form-data; name="field1"

one A section
--foo
Content-Disposition: form-data; name="userfile"; filename="songwriting"
Content-Type: text/plain
Content-Transfer-Encoding: binary

And another
--foo--
`)

	body = bytes.NewBuffer(postdata)
	req, err = http.NewRequest("POST", "http://qing.wps.cn/postfile", body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.Header.Set("Content-Type", "multipart/form-data; boundary=foo")

	router.ServeHTTP(nil, req)

	// check multipart
	if len(context.Multipart) != 2 {
		t.Fatal("context.Multipart len != 2", len(context.Multipart))
	}

	part := context.Multipart[0]
	if part.FormName != "field1" {
		t.Fatal("part.FormName != field1")
	}
	if string(part.Data) != "one A section" {
		t.Fatal("string(part.Data) != one A section")
	}

	part = context.Multipart[1]
	if part.FormName != "userfile" {
		t.Fatal("part.FormName != userfile")
	}
	if part.FileName != "songwriting" {
		t.Fatal("part.FileName != songwriting")
	}
	if string(part.Data) != "And another" {
		t.Fatal("string(part.Data) != And another")
	}

	if _, ok := context.Values["field1"]; !ok {
		t.Fatal("field1 not in Values")
	}
	if _, ok := context.Values["userfile"]; ok {
		t.Fatal("userfile should not be in Values")
	}
}
