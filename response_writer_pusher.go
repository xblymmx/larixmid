package larixmid

import (
	"net/http"
	"errors"
)

func (r *responseWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := r.ResponseWriter.(http.Pusher)
	if ok {
		return pusher.Push(target, opts)
	}

	return errors.New("response does not support http push")
}
