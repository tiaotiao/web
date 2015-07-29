package web

import (
	"path"
)

// A router to handle url and to manage middlewares.
// Web is a router based on the root path.
//
// Example:
//
// 	w := web.NewWeb()              // w based on /
//
// 	r1 := w.SubRouter("/api")      // r1 based on /api
// 	r2 := r1.SubRouter("message")  // r2 based on /api/message
//	r2.Append(NewAuthMiddleware()) // r2 add AuthMiddleware
//	r3 := r2.SubRouter("/")        // r3 based on /api/message, with AuthMiddleware. The same as r2.
//
//	r2.Append(NewRateLimitMiddleware())     // r2 add RateLimitMiddleware
//
//	r1.Handle("GET", "/status", Status)     // GET    /api/status, without Middleware
//	r2.Handle("POST", "/add", AddMessage)   // POST   /api/message/add, with AuthMiddleware and RateLimitMiddleware
//	r3.Handle("DELETE", "/del", DelMessage) // DELETE /api/message/del, with AuthMiddleware
//
type Router interface {
	// Register the Handler to handle this url. Method can be http methods like "GET", "POST",
	// "DELETE" etc, case insensitive. The path is related to the base path of this router.
	// All middlewares already in this router will be applied to this handler. But new
	// middlewares after will not affect. It will panic if you handle two functions with
	// the same url.
	Handle(method string, path string, fn Handler) *MiddlewaresManager

	// Append a middleware to this router. Middlewares will applied to handler in sequence.
	Append(midd Middleware)

	// Get a sub router with add this path. Note that the base path of sub router
	// is based on current base path. Middlewares in the sub router is a copy of
	// this router. But after this, they will be independent with each other.
	SubRouter(path string) Router
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

func (r *router) Append(midd Middleware) {
	r.midwares.Append(midd)
}

func (r *router) Handle(method string, urlpath string, fn Handler) *MiddlewaresManager {
	midwares := r.midwares.duplicate() // copy one
	urlpath = path.Join(r.base, urlpath)

	r.web.handle(method, urlpath, fn, midwares)

	return midwares
}

func (r *router) SubRouter(basePath string) Router {
	base := path.Join(r.base, basePath)
	midwares := r.midwares.duplicate()
	return newRouter(r.web, base, midwares)
}

var _ Router = (*router)(nil)
