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
	loggerFn LoggerFn
}

func NewProcessor(
	loggerFn LoggerFn,
) Processor {
	return Processor{
		loggerFn: loggerFn,
	}
}

func (p Processor) ProcessOutput(
	w http.ResponseWriter,
	r *http.Request,
	out any,
	outError error,
	statusCode int,
) error {
	output, err := output.JSON(
		r.Context(),
		w,
		r,
		out,
		outError,
		statusCode,
	)
	if err != nil {
		p.loggerFn(r).Error(
			"Error handling output JSON: %s",
			err,
		)

		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	p.loggerFn(r).Trace("Client output", output)

	return nil
}
