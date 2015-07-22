package web

import (
	"path"
	"reflect"
	"strings"
)

type Router interface {
	Append(midd Middleware) *MiddlewaresManager
	Handle(path string, handler interface{}) *MiddlewaresManager
	HandleFunc(method string, path string, fn WebFunc) *MiddlewaresManager
	SubRouter(pathPerfix string) Router
	Clear()
}

type router struct {
	web      *Web
	base     string
	midwares *MiddlewaresManager
}

func newRouter(web *Web, basePath string, midwares *MiddlewaresManager) *router {
	if midwares == nil {
		midwares = newMiddlewaresManager()
	}
	r := new(router)
	r.web = web
	r.base = basePath
	r.midwares = midwares
	return r
}

func (r *router) Append(midd Middleware) *MiddlewaresManager {
	return r.midwares.Append(midd)
}

func (r *router) Clear() {
	r.midwares.Clear()
}

func (r *router) Handle(urlpath string, handler interface{}) *MiddlewaresManager {
	midwares := r.midwares.Duplicate() // copy one
	urlpath = path.Join(r.base, urlpath)

	v := reflect.ValueOf(handler)

	if v.NumMethod() <= 0 {
		panic("handler has no method")
	}

	properCase := func(s string) string {
		if s == "" {
			return ""
		}
		a := strings.ToUpper(s[:1])
		b := strings.ToLower(s[1:])
		return a + b
	}

	var ok bool

	for _, s := range ALL_METHODS {
		mv := v.MethodByName(properCase(s))
		if !mv.IsValid() {
			continue
		}

		r.web.handle(s, urlpath, mv.Interface(), midwares)
		ok = true
	}

	if !ok {
		panic("handler has no valid method")
	}

	return midwares
}

func (r *router) HandleFunc(method string, urlpath string, fn WebFunc) *MiddlewaresManager {
	midwares := r.midwares.Duplicate() // copy one
	urlpath = path.Join(r.base, urlpath)

	r.web.handle(method, urlpath, fn, midwares)

	return midwares
}

func (r *router) SubRouter(pathPerfix string) Router {
	base := path.Join(r.base, pathPerfix)
	midwares := r.midwares.Duplicate()
	return newRouter(r.web, base, midwares)
}

///////////////////////////////////////////////////////////////////////////////

const (
	POST    = "POST"
	GET     = "GET"
	DELETE  = "DELETE"
	PUT     = "PUT"
	OPTIONS = "OPTIONS"
)

var (
	ALL_METHODS = []string{GET, POST, DELETE, PUT, OPTIONS}
)
