package web

import (
	"io/ioutil"
	"net/http"
	"net/textproto"
	"sync/atomic"
)

var MaxBodyLength int64 = 20 * (1 << 20) // 20M

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

func newContext(w http.ResponseWriter, r *http.Request) (*Context, error) {
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

var globalReqId int64
