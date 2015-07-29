package web

import (
	"io/ioutil"
	"net/http"
	"net/textproto"
	"strconv"
	"sync/atomic"
)

const MaxBodyLength int64 = 20 * (1 << 20) // 20M

var globalReqId int64

type Context struct {
	Request   *http.Request
	RequestId int64

	ResponseWriter http.ResponseWriter
	ResponseHeader http.Header

	Values map[string]interface{}

	RawPostData []byte

	Multipart []*Part
}

type Part struct {
	FormName string
	FileName string
	Header   textproto.MIMEHeader
	Data     []byte
}

func NewContext(w http.ResponseWriter, r *http.Request) (*Context, error) {
	var err error

	c := new(Context)

	c.Request = r
	c.RequestId = atomic.AddInt64(&globalReqId, 1)

	if w != nil {
		c.ResponseHeader = w.Header()
	}
	c.ResponseWriter = w

	c.Values = make(map[string]interface{})

	if r.Body != nil {
		mr := http.MaxBytesReader(w, r.Body, MaxBodyLength)

		c.RawPostData, err = ioutil.ReadAll(mr)
		r.Body.Close()

		if err != nil {
			return c, NewError(err.Error(), StatusBadRequest)
		}
	}

	return c, nil
}

func (c *Context) Set(name string, val interface{}) {
	c.Values[name] = val
}

func (c *Context) Get(name string) (interface{}, bool) {
	v, ok := c.Values[name]
	if ok {
		return v, true
	}

	s := c.Request.FormValue(name)
	if s != "" {
		return s, true
	}

	v, ok = c.Request.Form[name]
	if ok {
		return v, true
	}

	return nil, false
}

func (c *Context) GetInt(name string) (int, bool) {
	v, ok := c.Get(name)
	if !ok {
		return 0, false
	}

	switch x := v.(type) {
	case int:
		return x, true
	case string:
		i, err := strconv.Atoi(x)
		if err != nil {
			return 0, false
		}
		return i, true
	}

	return 0, false
}

func (c *Context) GetInt64(name string) (int64, bool) {
	v, ok := c.Get(name)
	if !ok {
		return 0, false
	}

	switch x := v.(type) {
	case int64:
		return x, true
	case string:
		i, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			return 0, false
		}
		return i, true
	}

	return 0, false
}

func (c *Context) GetString(name string) (string, bool) {
	v, ok := c.Get(name)
	if !ok {
		return "", false
	}

	s, ok := v.(string)
	if !ok {
		return "", false
	}

	return s, true
}
