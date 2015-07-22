package web

import (
	"encoding/json"
	"net/http"
)

type StatusCode interface {
	StatusCode() int
}

type Writeable interface {
	OnWrite(w http.ResponseWriter) error
}

///////////////////////////////////////////////////////////////////////////////

type Responser interface {
	Response(c *Context, result interface{}) (code int, err error)
}

///////////////////////////////////////////////////////////////////////////////

type DefaultResponser struct {
}

func NewDefaultResponser() *DefaultResponser {
	return new(DefaultResponser)
}

func (r *DefaultResponser) Response(c *Context, result interface{}) (int, error) {
	if result == nil {
		return StatusOK, nil
	}
	var err error
	w := c.ResponseWriter

	switch v := result.(type) {
	case []byte:
		_, err := w.Write(v)
		return StatusOK, err

	case string:
		_, err := w.Write([]byte(v))
		return StatusOK, err

	case error: // unknown error
		if _, ok := v.(*Error); ok {
			break
		}
		result = NewErrorMsg("server error", v.Error(), StatusInternalServerError)
	}

	// get status code
	var code int = StatusOK
	sc, ok := result.(StatusCode)
	if ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)

	// write data
	err = r.writeResult(w, result)
	if err != nil {
		return StatusInternalServerError, err
	}

	return code, nil
}

func (r *DefaultResponser) writeResult(w http.ResponseWriter, result interface{}) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	if _, err = w.Write(data); err != nil {
		return err
	}
	return nil
}
