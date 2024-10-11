package outputprocessor

import (
	"net/http"

	"github.com/pakkasys/fluidapi-extended/output"
)

type ILogger interface {
	Trace(messages ...any)
	Error(messages ...any)
}

type LoggerFn func(r *http.Request) ILogger

type Processor struct {
	LoggerFn LoggerFn
}

func (p Processor) ProcessOutput(
	w http.ResponseWriter,
	r *http.Request,
	out any,
	outError error,
	status int,
) error {
	output, err := output.JSON(r.Context(), w, r, out, outError, status)
	if err != nil {
		if p.LoggerFn != nil {
			p.LoggerFn(r).Error("Error handling output JSON: %s", err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	if p.LoggerFn != nil {
		p.LoggerFn(r).Trace("Client output", output)
	}
	return nil
}
