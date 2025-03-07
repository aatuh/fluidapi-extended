package inputlogic

import (
	"net/http"

	"github.com/pakkasys/fluidapi/core"
)

// Callback is the function that handles the request.
type Callback func(w http.ResponseWriter, r *http.Request)

// Middleware constructs a new middleware that calls the provided callback
//
//   - callback: The function that handles the request.
func Middleware(callback Callback) core.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callback(w, r)
			next.ServeHTTP(w, r)
		})
	}
}
