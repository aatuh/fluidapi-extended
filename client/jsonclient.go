package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Constants for HTTP headers and content types.
const (
	contentTypeHeader = "Content-Type"
	applicationJSON   = "application/json"
)

// Response represents the response from a client request.
type Response[Output any] struct {
	Response *http.Response // The HTTP response object.
	Output   *Output        // The output data of the API response.
}

// SendOptions represents options for sending a request.
type SendOptions struct {
	Headers map[string]string
	Cookies []http.Cookie
	Body    map[string]any
}

// JSONClient represents a JSON API client with a base URL and a custom HTTP
// client.
type JSONClient[Output any] struct {
	baseURL    string
	httpClient *http.Client
}

// NewJSONClient creates a new Client instance with the given base URL and timeout.
// If timeout is zero or negative, a default of 10 seconds is used.
//
// Parameters:
//   - baseURL: The base URL of the service.
//   - timeout: The timeout duration for HTTP requests.
//
// Returns:
//   - *Client: A new JSONClient instance.
func NewJSONClient[Output any](
	baseURL string, timeout time.Duration,
) *JSONClient[Output] {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &JSONClient[Output]{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Send sends a request to the specified URL with the provided input, host, and
// HTTP method and returns a Response containing the output, and HTTP response.
// It will set the Content-Type header to application/json if the header is not
// set.
//
// Parameters:
//   - host: The host part of the URL to send the request to.
//   - url: The endpoint URL path and query parameters.
//   - method: The HTTP method (e.g., GET, POST).
//   - sendOptions: An optional SendOptions struct to configure the request.
//
// Returns:
//   - Response: A Response containing the output and HTTP response.
//   - error: An error if the request fails.
func (c *JSONClient[Output]) Send(
	ctx context.Context, url string, method string, sendOptions *SendOptions,
) (*Response[Output], error) {
	var useSendOptions *SendOptions
	if sendOptions == nil {
		useSendOptions = &SendOptions{}
	} else {
		useSendOptions = sendOptions
	}
	if useSendOptions.Body != nil && method == http.MethodGet {
		return nil, fmt.Errorf("SendRequest: GET request must not have body")
	}

	// Marshal the body into a JSON reader.
	bodyReader, err := marshalBody(useSendOptions.Body)
	if err != nil {
		return nil, fmt.Errorf("SendRequest: body marshal error: %v", err)
	}

	// Set the Content-Type header if it is not set.
	if useSendOptions.Headers != nil &&
		useSendOptions.Headers[contentTypeHeader] == "" {
		useSendOptions.Headers[contentTypeHeader] = applicationJSON
	}

	// Create the request.
	req, err := createRequest(
		ctx,
		method,
		fmt.Sprintf("%s%s", c.baseURL, url),
		bodyReader,
		useSendOptions.Headers,
		useSendOptions.Cookies,
	)
	if err != nil {
		return nil, fmt.Errorf("SendRequest: error creating request: %v", err)
	}

	// Send the request.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SendRequest: request error: %v", err)
	}

	// Unmarshal the response body into the output object.
	output, err := responseToPayload(resp, new(Output))
	if err != nil {
		return nil, fmt.Errorf("SendRequest: response unmarshal error: %v", err)
	}
	return &Response[Output]{
		Response: resp,
		Output:   output,
	}, nil
}

// createRequest creates a new request with the specified request data.
func createRequest(
	ctx context.Context,
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
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("createRequest: error creating request: %v", err)
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
		return nil, fmt.Errorf("marshalBody: JSON marshal error: %v", err)
	}
	return bytes.NewReader(bodyBytes), nil
}

// responseToPayload unmarshals the response body into the output object.
func responseToPayload[T any](r *http.Response, output *T) (*T, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("responseToPayload: read error: %v", err)
	}

	// Handle non-JSON responses.
	ct := r.Header.Get(contentTypeHeader)
	if !strings.HasPrefix(ct, applicationJSON) {
		return nil, fmt.Errorf(
			"responseToPayload: expected JSON response, got Content-Type: %s, body: %s",
			ct,
			string(body),
		)
	}

	// For JSON responses, attempt to unmarshal.
	if err := json.Unmarshal(body, output); err != nil {
		return nil, fmt.Errorf(
			"responseToPayload: JSON unmarshal error: %v, body: %s",
			err,
			string(body),
		)
	}
	return output, nil
}
