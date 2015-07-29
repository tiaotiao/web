package web

import (
	"fmt"
	"runtime/debug"
)

type Middleware interface {
	ServeMiddleware(c *Context) error
}

type ResponseMiddleware interface {
	ServeResponse(c *Context, result interface{}) (interface{}, error)
}

///////////////////////////////////////////////////////////////////////////////

type middlewaresManager struct {
	midds []Middleware
}

func newMiddlewaresManager() *middlewaresManager {
	m := new(middlewaresManager)
	m.midds = make([]Middleware, 0, 8)
	return m
}

func (m *middlewaresManager) Append(midd Middleware) *middlewaresManager {
	if midd == nil {
		return m
	}
	m.midds = append(m.midds, midd)
	return m
}

func (m *middlewaresManager) Duplicate() *middlewaresManager {
	d := newMiddlewaresManager()
	copy(d.midds, m.midds)
	return d
}

func (m *middlewaresManager) ServeMiddlewares(c *Context) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = NewErrorMsg("server error", fmt.Sprintf("Panic: %v\n%v", e, debug.Stack()), StatusInternalServerError)
		}
	}()

	for _, midd := range m.midds {
		err = midd.ServeMiddleware(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *middlewaresManager) ServeResponses(c *Context, r interface{}) (rt interface{}) {
	defer func() {
		if e := recover(); e != nil {
			rt = NewErrorMsg("server error", fmt.Sprintf("Panic: %v", e, debug.Stack()), StatusInternalServerError)
		}
	}()

	var err error
	for _, midd := range m.midds {
		if respProcessor, ok := midd.(ResponseMiddleware); ok {
			r, err = respProcessor.ServeResponse(c, r)
			if err != nil {
				return err
			}
		}
	}
	return r
}
