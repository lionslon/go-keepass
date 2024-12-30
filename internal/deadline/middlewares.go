package deadline

import (
	"context"
	"net/http"
	"time"
)

const (
	contextTimeOut = 2 * time.Second
)

// Middleware Устанавливаем deadline для контекста
func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if _, ok := r.Context().Deadline(); !ok {
			ctx, cancel := context.WithTimeout(r.Context(), contextTimeOut)
			defer cancel()
			h.ServeHTTP(w, r.WithContext(ctx))
		} else {
			h.ServeHTTP(w, r)
		}
	})
}
