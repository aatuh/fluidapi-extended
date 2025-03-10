package cors

import (
	"net/http"
	"slices"
	"strings"

	"github.com/pakkasys/fluidapi/core"
)

// CORSOptions encapsulates configuration options for the CORS middleware.
type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// Middleware creates a new CORS middleware based on CORSOptions.
//
// The middleware inspects the incoming request's "Origin" header and, if the
// origin is in the allowed list, it sets the appropriate CORS headers on the
// response. When AllowCredentials is enabled, the middleware echoes the
// request’s origin rather than using the wildcard "*" to comply with the CORS
// spec. Additionally, it sets the allowed HTTP methods and headers as
// configured.
//
// Parameters:
//   - opts: CORSOptions containing configuration for allowed origins, methods,
//     headers, and whether credentials are permitted.
//
// Returns:
//   - core.Middleware: A middleware function that applies the CORS
//     configuration.
func Middleware(opts CORSOptions) core.Middleware {
	// Check if "*" is allowed.
	isWildcard := slices.Contains(opts.AllowedOrigins, "*")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// If credentials are allowed, don't use "*" even if configured.
			if opts.AllowCredentials && origin != "" &&
				slices.Contains(opts.AllowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if isWildcard {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if slices.Contains(opts.AllowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set(
				"Access-Control-Allow-Methods",
				strings.Join(opts.AllowedMethods, ","),
			)
			// Combine default headers with user-specified ones.
			allHeaders := slices.Concat(
				[]string{"Content-Type"}, opts.AllowedHeaders,
			)
			w.Header().Set(
				"Access-Control-Allow-Headers", strings.Join(allHeaders, ","),
			)

			// Only set credentials header if allowed.
			if opts.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			next.ServeHTTP(w, r)
		})
	}
}
