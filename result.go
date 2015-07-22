package web

import (
	"fmt"
	"net/http"
)

var ResultOK = Result{"result": "ok"}

///////////////////////////////////////////////////////////////////////////////

type Result map[string]interface{}

func (r Result) StatusCode() int {
	v, ok := r["__code__"]
	if !ok {
		return StatusOK
	}
	code, ok := v.(int)
	if !ok {
		return StatusOK
	}
	delete(r, "__code__")
	return code
}

func (r Result) SetStatusCode(code int) Result {
	r["__code__"] = code
	return r
}

var _ StatusCode = (*Result)(nil)

///////////////////////////////////////////////////////////////////////////////

type Message struct {
	Message string `json:"message"`
	Code    int    `json:"-"`
}

func NewMessage(msg string, code int) *Message {
	e := &Message{Message: msg, Code: code}
	return e
}

func (e *Message) StatusCode() int {
	return e.Code
}

var _ StatusCode = (*Message)(nil)

///////////////////////////////////////////////////////////////////////////////

type Error struct {
	Err     string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"-"`
}

func NewError(e string, code int) *Error {
	return NewErrorMsg(e, "", code)
}

func NewErrorMsg(e, msg string, code int) *Error {
	err := new(Error)
	err.Err = e
	err.Message = msg
	err.Code = code
	return err
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v", e)
}

func (e *Error) StatusCode() int {
	return e.Code
}

var _ error = (*Error)(nil)
var _ StatusCode = (*Error)(nil)

///////////////////////////////////////////////////////////////////////////////

const (
	StatusOK                  = http.StatusOK                  // 200
	StatusBadRequest          = http.StatusBadRequest          // 400
	StatusUnauthorized        = http.StatusUnauthorized        // 401
	StatusForbidden           = http.StatusForbidden           // 403
	StatusNotFound            = http.StatusNotFound            // 404
	StatusInternalServerError = http.StatusInternalServerError // 500
)
