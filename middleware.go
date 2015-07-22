package web

import (
	"fmt"
	"runtime/debug"
)

type Middleware interface {
	Name() string
	ServeMiddleware(c *Context) error
}

type ResponseProcessor interface {
	ProcessResponse(c *Context, result interface{}) (interface{}, error)
}

///////////////////////////////////////////////////////////////////////////////

// Middleware define in middleware_interface.go
type MiddlewaresManager struct {
	midds []Middleware
}

func newMiddlewaresManager() *MiddlewaresManager {
	m := new(MiddlewaresManager)
	m.midds = make([]Middleware, 0, 8)
	return m
}

func (m *MiddlewaresManager) Append(midd Middleware) *MiddlewaresManager {
	if midd == nil {
		return m
	}
	m.midds = append(m.midds, midd)
	return m
}

func (m *MiddlewaresManager) Clear() {
	m.midds = make([]Middleware, 0, 8)
}

func (m *MiddlewaresManager) Remove(name string) bool {
	var ok bool
	for i, midd := range m.midds {
		if midd.Name() == name {
			m.midds = removeFromList(m.midds, i)
			ok = true
			break
		}
	}
	return ok
}

func removeFromList(midds []Middleware, x int) []Middleware {
	if x < 0 || len(midds) <= x {
		return midds
	}
	midds[x] = nil
	for i := x + 1; i < len(midds); i++ {
		midds[i-1] = midds[i]
		midds[i] = nil
	}
	return midds[:len(midds)-1]
}

func (m *MiddlewaresManager) Duplicate() *MiddlewaresManager {
	d := newMiddlewaresManager()
	copy(d.midds, m.midds)
	return d
}

func (m *MiddlewaresManager) ServeMiddlewares(c *Context) (err error) {
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

func (m *MiddlewaresManager) ProcessResponse(c *Context, r interface{}) (rt interface{}) {
	defer func() {
		if e := recover(); e != nil {
			rt = NewErrorMsg("server error", fmt.Sprintf("Panic: %v", e, debug.Stack()), StatusInternalServerError)
		}
	}()

	var err error
	for _, midd := range m.midds {
		if respProcessor, ok := midd.(ResponseProcessor); ok {
			r, err = respProcessor.ProcessResponse(c, r)
			if err != nil {
				return err
			}
		}
	}
	return r
}