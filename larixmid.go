package larixmid

import (
	"net/http"
	"log"
	"os"
)

const (
	DefaultAddr = ":8080"
)

type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)

func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	h(w, r, next)
}

// implements net/http.Handler
type middleware struct {
	handler Handler
	next    *middleware
}

// implements net/http.Handler
func (m middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.handler.ServeHTTP(w, r, m.next.ServeHTTP)
}

// convert http.Handler to larixmid Handler
func Wrap(handler http.Handler) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		handler.ServeHTTP(w, r)
		next(w, r)
	})
}

// convert http.HandlerFunc to larixmid Handler
func WrapFunc(handleFunc http.HandlerFunc) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		handleFunc(w, r)
		next(w, r)
	})
}

type Larixmid struct {
	middleware middleware
	handlers   []Handler
}

func New(handlers ...Handler) *Larixmid {
	return &Larixmid{
		handlers:   handlers,
		middleware: build(handlers),
	}
}

// Larixmid is a stack of Middleware Handlers that can be invoked as an http.Handler.
// Larixmid middleware is evaluated in the order that they are added to the stack using the Use and UseHandler methods.
func (la *Larixmid) With(handlers ...Handler) *Larixmid {
	return New(
		append(la.handlers, handlers...)...
	)
}

// Recovery - Panic Recovery middleware
// Logger - Request/Response Logging
// Static - Static File Serving
func Classic() *Larixmid {
	// todo NewRecovery()
	return New(NewLogger(), NewStatic(http.Dir("public")), NewRecovery())
}

// net/http.Handler
func (la *Larixmid) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	la.middleware.ServeHTTP(NewResponseWriter(w), r)
}

// adds larixmid style Handler
func (la *Larixmid) Use(handler Handler) {
	if handler == nil {
		panic("handler cannot be nil")
	}
	la.handlers = append(la.handlers, handler)
	la.middleware = build(la.handlers)
}

// adds larixmid style HandlerFunc
func (la *Larixmid) UseFunc(fn func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)) {
	la.Use(HandlerFunc(fn))
}

// adds http.Handler style Handler
func (la *Larixmid) UseHandler(handler http.Handler) {
	la.Use(Wrap(handler))
}

// add http.HandleFunc style handler to middleware stack
func (la *Larixmid) UseHandlerFunc(f func(w http.ResponseWriter, r *http.Request)) {
	la.UseHandler(http.HandlerFunc(f))
}

func (la *Larixmid) Run(addr ...string) {
	logger := log.New(os.Stdout, "[Larixmid]", 0)
	finalAddr := detectAddr(addr...)
	logger.Println("listening on", finalAddr)
	logger.Fatalln(http.ListenAndServe(finalAddr, la))
}

func detectAddr(addr ...string) string {
	if len(addr) > 0 {
		return addr[0]
	}
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return DefaultAddr
}

func (la *Larixmid) Handlers() []Handler {
	return la.handlers
}

func build(handlers []Handler) middleware {
	var next middleware

	if len(handlers) == 0 {
		return voidMiddleware()
	} else if len(handlers) > 1 {
		next = build(handlers[1:])
	} else { // last middleware
		next = voidMiddleware()
	}

	return middleware{handler: handlers[0], next: &next}
}

func voidMiddleware() middleware {
	return middleware{
		handler: HandlerFunc(func(w http.ResponseWriter, r *http.Request, handlerFunc http.HandlerFunc) {}),
		next:    &middleware{},
	}
}
