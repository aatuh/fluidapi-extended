package cors

import (
	"net/http"
	"slices"
	"strings"

	"github.com/pakkasys/fluidapi/core"
)

const (
	headerAllowOrigin      = "Access-Control-Allow-Origin"
	headerAllowMethods     = "Access-Control-Allow-Methods"
	headerAllowHeaders     = "Access-Control-Allow-Headers"
	headerAllowCredentials = "Access-Control-Allow-Credentials"

	originHeader = "Origin"
)

// TODO: Opinionated
var (
	corsAllowHeaders = []string{"Content-Type"}
)

// Middleware creates a new CORS middleware
//
//   - allowedOrigins: The list of allowed origins
//   - allowedMethods: The list of allowed methods
//   - allowedHeaders: The list of allowed headers
func Middleware(
	allowedOrigins []string,
	allowedMethods []string,
	allowedHeaders []string,
) core.Middleware {
	return func(next http.Handler) http.Handler {
		isWildcardOrigin := slices.Contains(allowedOrigins, "*")

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get(originHeader)
			if isWildcardOrigin {
				w.Header().Set(headerAllowOrigin, "*")
			} else if slices.Contains(allowedOrigins, origin) {
				w.Header().Set(headerAllowOrigin, origin)
			}

			w.Header().Set(
				headerAllowMethods,
				strings.Join(allowedMethods, ","),
			)

			w.Header().Set(
				headerAllowHeaders,
				strings.Join(
					slices.Concat(corsAllowHeaders, allowedHeaders), ",",
				),
			)

			w.Header().Set(headerAllowCredentials, "true")

			next.ServeHTTP(w, r)
		})
	}
}
