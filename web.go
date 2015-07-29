package web

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// Web is the main object of the framework. All things start from here.
// To create an Web object, call NewWeb().
//
// You can use this web framework to handle HTTP requests by writing your own functions
// and return results with JSON format (or other custom format if you want).
//
// Note that this framework is designed for API server. To keep it simple, it doesn't support
// static page, html template, css, javascript etc.
//
// Here is an example:
//
// 	import "github.com/tiaotiao/web"
//
// 	func Hello(c *web.Context) interface{} {
// 		return "hello world" // return a string will be write directly with out processing.
// 	}
//
// 	func Echo(c *web.Context) interface{} {
// 		args := struct {
// 			Message string `web:"message,required"`
// 		}{}
//
//		// scheme args from c.Values into args by the indication of struct tags.
// 		if err := web.Scheme(c.Values, &args); err != nil {
// 			return err 	// error occured if lack of argument or wrong type.
// 		}
//
// 		return web.Result{"message": args.Message} // a map or struct will be formated to JSON
// 	}
//
// 	func main() {
// 		w := web.NewWeb()
//
// 		w.Handle("GET", "/api/hello", Hello)
// 		w.Handle("GET", "/api/echo", Echo)
//
// 		w.ListenAndServe("tcp", ":8082")  // block forever until be closed
// 	}
//
type Web struct {
	mux    *mux.Router
	router *router

	handlers map[string]*WebHandler

	listeners []net.Listener

	responser Responser
	logger    Logger
	stat      *Stat

	wg     sync.WaitGroup
	closed bool
}

// Create an Web object.
func NewWeb() *Web {
	w := new(Web)

	w.mux = mux.NewRouter().StrictSlash(true)

	w.router = newRouter(w, "/", nil)

	w.handlers = make(map[string]*WebHandler, 128)

	w.responser = new(DefaultResponser)

	w.stat = newStat()
	w.closed = false
	return w
}

// ServeHTTP used for implements http.Handler interface. No need to be called by user.
func (w *Web) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	w.wg.Add(1)
	if !w.closed {
		w.stat.onServe(req)

		w.mux.ServeHTTP(rw, req)

		w.stat.onDone(req)
	}
	w.wg.Done()
}

// Listen an address and start to serve. Blocked until be closed or some error occurs.
// See Web.Listen and Web.Serve.
func (w *Web) ListenAndServe(protocol string, addr string) error {
	var err error

	err = w.Listen(protocol, addr)
	if err != nil {
		return err
	}

	return w.Serve()
}

// Listen on the local network address. Argument protocol usually be 'tcp' or 'unix'.
// Argument addr usually be the format as 'host:port'. For example:
//		Listen("tcp", ":8080")	// listen on 8080 port
//		Listen("tcp", "google.com:http") // listen on 80 port from google.com
//		Listen("unix", "/var/run/web_server.sock")
// See net.Listen for more.
func (w *Web) Listen(protocol string, addr string) error {
	var err error

	l, err := net.Listen(protocol, addr)
	if err != nil {
		return err
	}

	w.listeners = append(w.listeners, l)

	return nil
}

// Start to serve all listeners. It will block until all listeners closed.
func (w *Web) Serve() error {
	var err error
	if len(w.listeners) == 0 {
		return fmt.Errorf("not listening")
	}

	wg := sync.WaitGroup{}

	serve := func(l net.Listener) {
		svrMux := http.NewServeMux()
		svrMux.Handle("/", w)

		svr := http.Server{Handler: svrMux, ReadTimeout: HttpReadTimeout}

		err = svr.Serve(l)

		l.Close()

		wg.Done()
	}

	for _, l := range w.listeners {
		wg.Add(1)
		go serve(l)
	}

	wg.Wait()

	return err
}

// Close all listeners and stop serve HTTP.
func (w *Web) Close() {
	for _, l := range w.listeners {
		l.Close()
	}

	w.closed = true

	w.wg.Wait()
}

// Register an WebFunc as a handler for this url. See Router.Handle
func (w *Web) Handle(method string, path string, fn WebFunc) *MiddlewaresManager {
	return w.router.Handle(method, path, fn)
}

// Get a sub router with the perfix path. See Router.SubRouter
func (w *Web) SubRouter(pathPerfix string) Router {
	return w.router.SubRouter(pathPerfix)
}

// Append a middleware. See Router.Append
func (w *Web) Append(midd Middleware) {
	w.router.Append(midd)
}

// To setup a custom responser to process the result which returned from WebFunc and then to write into response body.
// The responser must implements the Responser interface.
//
// With out doing anything, the DefaultResponser will write string and []byte directly or write map, struct and
// other types in JSON format. See DefaultResponser for more detail.
func (w *Web) SetResponser(r Responser) {
	w.responser = r
}

// Set a logger to track request logs.
func (w *Web) SetLogger(l Logger) {
	w.logger = l
}

// Get statistic.
func (w *Web) GetStatistic() *Stat {
	return w.stat
}

// Get all registed handlers.
func (w *Web) GetHandlers() map[string]*WebHandler {
	return w.handlers
}

func (w *Web) handle(method, urlpath string, fn WebFunc, midwares *MiddlewaresManager) {
	var h *WebHandler

	h = newWebHandler(fn, midwares, w.responser, w.logger)

	// match prefix
	var prefix bool
	if strings.HasSuffix(urlpath, "*") {
		urlpath = strings.TrimSuffix(urlpath, "*")
		prefix = true
	}

	// register mux route
	var rt *mux.Route
	if prefix {
		rt = w.mux.PathPrefix(urlpath).Handler(h)
	} else {
		rt = w.mux.Handle(urlpath, h)
	}
	rt.Methods(strings.ToUpper(method))

	// add to map
	url := methodUrl(method, urlpath)
	_, ok := w.handlers[url]
	if ok {
		panic("url conflict: " + url)
	}
	w.handlers[url] = h

	h.stat.Path = url
	w.stat.Handlers = append(w.stat.Handlers, h.stat)
	return
}

func methodUrl(method string, path string) string {
	return method + " " + strings.ToLower(path)
}

var HttpReadTimeout = time.Minute

const (
	POST    = "POST"
	GET     = "GET"
	DELETE  = "DELETE"
	PUT     = "PUT"
	OPTIONS = "OPTIONS"
)
