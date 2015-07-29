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

var HttpReadTimeout = time.Minute

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

func NewWeb() *Web {
	w := new(Web)

	w.mux = mux.NewRouter().StrictSlash(true)

	w.router = newRouter(w, "/", nil)

	w.handlers = make(map[string]*WebHandler, 128)

	w.responser = NewDefaultResponser()

	w.stat = newStat()
	w.closed = false
	return w
}

func (w *Web) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	w.wg.Add(1)
	if !w.closed {
		w.stat.onServe(req)

		w.mux.ServeHTTP(rw, req)

		w.stat.onDone(req)
	}
	w.wg.Done()
}

func (w *Web) ListenAndServe(protocol string, addr string) error {
	var err error

	err = w.Listen(protocol, addr)
	if err != nil {
		return err
	}

	return w.Serve()
}

func (w *Web) Listen(protocol string, addr string) error {
	var err error

	l, err := net.Listen(protocol, addr)
	if err != nil {
		return err
	}

	w.listeners = append(w.listeners, l)

	return nil
}

// serve all listeners
// It will block until all listeners closed.
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

func (w *Web) Close() {
	for _, l := range w.listeners {
		l.Close()
	}

	w.closed = true

	w.wg.Wait()
}

func (w *Web) Append(midd Middleware) {
	w.router.Append(midd)
}

func (w *Web) Handle(method string, path string, fn WebFunc) *MiddlewaresManager {
	return w.router.Handle(method, path, fn)
}

func (w *Web) SubRouter(pathPerfix string) Router {
	return w.router.SubRouter(pathPerfix)
}

func (w *Web) SetResponser(r Responser) {
	w.responser = r
}

func (w *Web) SetLogger(l Logger) {
	w.logger = l
}

func (w *Web) GetStatistic() *Stat {
	return w.stat
}

func (w *Web) GetHandlers() map[string]*WebHandler {
	return w.handlers
}

func (w *Web) handle(method, urlpath string, fn WebFunc, midwares *MiddlewaresManager) {
	var h *WebHandler

	h = NewWebHandler(fn, midwares, w.responser, w.logger)

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

const (
	POST    = "POST"
	GET     = "GET"
	DELETE  = "DELETE"
	PUT     = "PUT"
	OPTIONS = "OPTIONS"
)
