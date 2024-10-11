package server

import (
	"fmt"
	"net/http"

	"github.com/pakkasys/fluidapi/core/api"
	"github.com/pakkasys/fluidapi/core/server"
)

type ILogger interface {
	Info(messages ...any)
	Error(messages ...any)
}

type LoggerFn func(r *http.Request) ILogger

func MustStart(endpoints []api.Endpoint, port int, loggerFn LoggerFn) {
	err := server.HTTPServer(
		server.DefaultHTTPServer(
			port,
			endpoints,
			func(r *http.Request) func(messages ...any) {
				return loggerFn(r).Info
			},
			func(r *http.Request) func(messages ...any) {
				return loggerFn(r).Error
			},
		),
	)
	if err != nil {
		panic(fmt.Sprintf("Server error: %s", err))
	}
}
