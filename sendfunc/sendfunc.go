package sendfunc

import (
	"errors"

	"github.com/pakkasys/fluidapi-extended/output"
	"github.com/pakkasys/fluidapi-extended/urlencoder"

	"github.com/pakkasys/fluidapi/core/client"
	"github.com/pakkasys/fluidapi/endpoint/runner"
)

type Client[I any, O any] *runner.Client[I, O, output.Output[O]]

// SendFunc returns a send function that wraps the API client.
// It returns an error if the API returns an error.
func SendFunc[I any, O any](
	url string,
	method string,
) func(*I, string) (*client.Response[I, output.Output[O]], error) {
	return func(
		input *I,
		host string,
	) (*client.Response[I, output.Output[O]], error) {
		apiResponse, err := client.Send[I, output.Output[O]](
			input,
			url,
			host,
			method,
			urlencoder.URLEncoder{},
		)
		if err != nil {
			return nil, err
		}

		apiErr := output.APIError(apiResponse)
		if apiErr != nil {
			return apiResponse, apiErr
		}

		return apiResponse, nil
	}
}

// SendAndExtractPayload sends a request and extract the payload from the
// output.Output wrapper.
func SendAndExtractPayload[I any, O any](
	sendFunc func(*I, string) (*client.Response[I, output.Output[O]], error),
	input *I,
	clientHost string,
) (*O, error) {
	apiResponse, err := sendFunc(input, clientHost)
	if err != nil {
		return nil, err
	}

	return apiResponse.Output.Payload, nil
}

// GetOne retrieves one entity from an API. If not found, it returns an error.
// O is the type of the `Output`, which contains the slice.
func GetOne[I any, O any, T any](
	sendFunc runner.SendFunc[I, output.Output[O]],
	input *I,
	clientHost string,
	notFoundError error,
	extractSliceFn func(*O) []T,
) (*T, error) {
	outputPayload, err := SendAndExtractPayload(sendFunc, input, clientHost)
	if err != nil {
		return nil, err
	}

	entities := extractSliceFn(outputPayload)
	if len(entities) == 0 {
		if notFoundError != nil {
			return nil, notFoundError
		}
		return nil, errors.New("no entities found")
	}

	return &entities[0], nil
}
