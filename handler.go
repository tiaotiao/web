package web

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
	"time"
)

type Handler interface{}

///////////////////////////////////////////////////////////////////////////////

type handler struct {
	fn Handler

	reflectFn      reflect.Value
	reflectArgType reflect.Type

	midds *MiddlewaresManager

	responser Responser
	logger    Logger
}

func newHandler(fn Handler, midds *MiddlewaresManager, responser Responser, logger Logger) *handler {
	h := new(handler)

	if fn == nil {
		panic("func is nil")
	}
	h.fn = fn

	h.midds = midds

	if responser == nil {
		panic("responser == nil")
	}
	h.responser = responser

	h.logger = logger

	err := h.validateHandler(fn)
	if err != nil {
		panic(err.Error())
	}

	return h
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	var result interface{}
	var start = time.Now()

	defer func() {
		used := time.Since(start)

		// response
		code, err := h.responser.Response(w, result)
		if err != nil {
			result = err
		}

		if h.logger != nil {
			h.logger.OnLog(r, start, used, code, result)
		}
	}()

	// new context
	c, err := newContext(w, r)
	if err != nil {
		result = err
		return
	}

	// parse params
	err = ParseParams(c)
	if err != nil {
		result = err
		return
	}

	result = h.serve(c) // serve
	return
}

func (h *handler) serve(c *Context) (result interface{}) {
	defer func() {
		if e := recover(); e != nil {
			s := fmt.Sprintf("Panic: %v\n %v", e, debug.Stack())
			result = NewErrorMsg("server error", s, StatusInternalServerError)
		}
	}()

	// serve middlewares
	err := h.midds.serveMiddlewares(c)
	if err != nil {
		return err
	}

	// call
	result = h.call(c)

	return h.midds.serveResponses(c, result)
}

func (h *handler) call(c *Context) (result interface{}) {
	var in = make([]reflect.Value, 0, 2)

	in = append(in, reflect.ValueOf(c))

	if h.reflectArgType != nil {
		arg := reflect.New(h.reflectArgType)

		err := Scheme(c.Values, arg.Interface()) // auto scheme
		if err != nil {
			return err
		}

		in = append(in, arg.Elem())
	}

	outs := h.reflectFn.Call(in)

	return outs[0].Interface()
}

func (h *handler) validateHandler(fn Handler) error {
	v := reflect.ValueOf(fn)
	t := reflect.TypeOf(fn)

	if t.Kind() != reflect.Func {
		return fmt.Errorf("not func type, %v", t.String())
	}

	if (t.NumIn() != 1 && t.NumIn() != 2) || t.NumOut() != 1 {
		return fmt.Errorf("invalid num of args, %v", t.String())
	}

	// the first arg must be *Context
	ctxArg := t.In(0)
	ctxType := reflect.TypeOf(new(Context))
	if ctxArg != ctxType {
		return fmt.Errorf("The first input arg must be *Context, %v", t.String())
	}

	// if there is the second arg, it must be a struct
	if t.NumIn() == 2 {
		arg := t.In(1)
		if arg.Kind() != reflect.Struct {
			return fmt.Errorf("The arg must be a struct, %v", t.String())
		}
		h.reflectArgType = arg
	}

	// output must be an interface{}
	out := t.Out(0)
	i := reflect.TypeOf(new(interface{})).Elem()
	if out != i {
		return fmt.Errorf("Output arg must be a interface{}, %v", t.String())
	}

	h.reflectFn = v

	return nil
}
