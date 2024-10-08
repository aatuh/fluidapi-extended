package output

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pakkasys/fluidapi/core/api"
	"github.com/pakkasys/fluidapi/core/client"
)

const (
	contentType     = "Content-Type"
	applicationJSON = "application/json"
)

var Error = api.NewError[any]("ERROR")

type Output[T any] struct {
	Payload *T              `json:"payload,omitempty"`
	Error   *api.Error[any] `json:"error,omitempty"`
}

// // Stringer
// func (ouput *Output[any]) String() string {
// 	jsonBytes, err := json.Marshal(ouput)
// 	if err != nil {
// 		return fmt.Sprintf("Error marshaling output: %s", err)
// 	}
// 	return string(jsonBytes)
// }

// HandleSendError checks if there is an error from the Send request and returns
// the appropriate response.
//   - output: The response from the Send function.
//   - err: The error returned from the Send function or API processing.
func HandleSendError[I any, O any](
	output *client.Response[I, Output[O]],
	err error,
) (*client.Response[I, Output[O]], error) {
	if err != nil {
		return output, err
	}
	if output.Output == nil {
		return output, nil
	}
	if output.Output.Error != nil {
		return output, output.Output.Error
	}
	return output, err
}

// APIPayload returns the payload of the API response if there are no errors.
// Otherwise, it returns nil.
func APIPayload[I any, O any](output *client.Response[I, Output[O]]) *O {
	if output == nil {
		return nil
	}
	if APIError(output) != nil {
		return nil
	}
	if output.Output != nil {
		return output.Output.Payload
	}
	return nil
}

// APIError returns the error of the API response from the output.
func APIError[I any, O any](
	output *client.Response[I, Output[O]],
) *api.Error[any] {
	if output == nil {
		return nil
	}
	if output.Output == nil {
		return nil
	}
	return output.Output.Error
}

func JSON(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	outputData any,
	outputError error,
	statusCode int,
) (*Output[any], error) {
	output := Output[any]{
		Payload: &outputData,
		Error:   handleError(outputError),
	}

	jsonData, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}

	w.Header().Set(contentType, applicationJSON)
	w.WriteHeader(statusCode)
	_, err = w.Write(jsonData)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func handleError(outputError error) *api.Error[any] {
	if outputError == nil {
		return nil
	}
	switch errType := outputError.(type) {
	case *api.Error[any]:
		return errType
	default:
		return Error
	}
}
