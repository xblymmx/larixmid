package larixmid

import (
	"net/http"
	"net"
	"bufio"
	"errors"
)

type ResponseWriter interface {
	http.Flusher
	http.ResponseWriter

	Status() int

	Written() bool

	Size() int

	Before(func(ResponseWriter))
}

type responseWriter struct {
	http.ResponseWriter

	status int
	size   int

	beforeFuncs []beforeFunc
}

type beforeFunc func(ResponseWriter)

func NewResponseWriter(w http.ResponseWriter) ResponseWriter {
	rw := &responseWriter{
		ResponseWriter: w,
	}

	if _, ok := rw.ResponseWriter.(responseWriterCloseNotifier); ok {
		return &responseWriterCloseNotifier{rw}
	}

	return rw
}

func (rw *responseWriter) WriteHeader(n int) {
	rw.status = n
	rw.callBefore()
	rw.ResponseWriter.WriteHeader(n)
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) Size() int {
	return rw.size
}

func (rw *responseWriter) Written() bool {
	return rw.status != 0
}

func (rw *responseWriter) Before(f func(w ResponseWriter)) {
	rw.beforeFuncs = append(rw.beforeFuncs, f)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.Written() {
		rw.WriteHeader(http.StatusOK)
	}

	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func (rw *responseWriter) Flush() {
	flusher, ok := rw.ResponseWriter.(http.Flusher)
	if ok {
		if !rw.Written() {
			rw.WriteHeader(http.StatusOK)
		}
		flusher.Flush()
	}
}

func (rw *responseWriter) callBefore() {
	for i := len(rw.beforeFuncs) - 1; i >= 0; i-- {
		rw.beforeFuncs[i](rw)
	}
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("response writer doesnt support hijacker interface")
	}

	return hijacker.Hijack()

}

type responseWriterCloseNotifier struct {
	*responseWriter
}

func (rw *responseWriterCloseNotifier) CloseNotify() <-chan bool {
	return rw.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
