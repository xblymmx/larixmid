package larixmid

import "net/http"


func NewRecover() HandlerFunc {
	return HandlerFunc(func (w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		// todo
		next(w, r)
	})
}
