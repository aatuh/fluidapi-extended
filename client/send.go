package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	contentTypeHeader = "Content-Type"
	applicationJSON   = "application/json"
)

// Response represents the response from a client request, including the HTTP
// response, input data, and output data.
type Response[Output any] struct {
	Response *http.Response // The HTTP response object.
	Output   *Output        // The output data of the API response.
}

type SendOptions struct {
	Headers map[string]string
	Cookies []http.Cookie
	Body    map[string]any
}

// Send sends a request to the specified URL with the provided input, host, and
// HTTP method and returns a Response containing the output, and HTTP response.
//
//   - host: The host part of the URL to send the request to.
//   - url: The endpoint URL path and query parameters.
//   - method: The HTTP method (e.g., GET, POST).
//   - headers: HTTP headers to include in the request.
//   - cookies: Cookies to include in the request.
//   - body: Request body data. If the method is GET and body is not nil,
//     an error is returned.
func Send[Output any](
	host string,
	url string,
	method string,
	sendOptions *SendOptions,
) (*Response[Output], error) {
	var useSendOptions *SendOptions
	if sendOptions == nil {
		useSendOptions = &SendOptions{}
	} else {
		useSendOptions = sendOptions
	}

	if useSendOptions.Body != nil && method == http.MethodGet {
		return nil, fmt.Errorf("body cannot be set for GET requests")
	}

	bodyReader, err := marshalBody(useSendOptions.Body)
	if err != nil {
		return nil, err
	}
	if useSendOptions.Headers != nil &&
		useSendOptions.Headers[contentTypeHeader] == "" {
		useSendOptions.Headers[contentTypeHeader] = applicationJSON
	}

	req, err := createRequest(
		method,
		fmt.Sprintf("%s%s", host, url),
		bodyReader,
		useSendOptions.Headers,
		useSendOptions.Cookies,
	)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	output, err := responseToPayload(resp, new(Output))
	if err != nil {
		return nil, err
	}

	return &Response[Output]{
		Response: resp,
		Output:   output,
	}, nil
}

// createRequest creates a new request with the specified method, URL, and body.
func createRequest(
	method string,
	url string,
	bodyReader io.Reader,
	headers map[string]string,
	cookies []http.Cookie,
) (*http.Request, error) {
	var body io.Reader
	if bodyReader != nil {
		body = bodyReader
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}
	for _, cookie := range cookies {
		req.AddCookie(&cookie)
	}

	return req, nil
}

// marshalBody marshals the body into a JSON reader.
func marshalBody(body any) (*bytes.Reader, error) {
	if body == nil {
		return bytes.NewReader(nil), nil
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(bodyBytes), nil
}

// responseToPayload unmarshals the response body into the output object.
func responseToPayload[T any](r *http.Response, output *T) (*T, error) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, output); err != nil {
		return nil, fmt.Errorf(
			"JSON unmarshal error: %v, body: %s", err, string(body),
		)
	}

	return output, nil
}
