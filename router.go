package web

import (
	"path"
)

type Router interface {
	Handle(method string, path string, fn WebFunc) *MiddlewaresManager

	Append(midd Middleware) *MiddlewaresManager
	SubRouter(pathPerfix string) Router
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

func (r *router) Handle(method string, urlpath string, fn WebFunc) *MiddlewaresManager {
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
