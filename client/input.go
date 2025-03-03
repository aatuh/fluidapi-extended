package client

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

const (
	sourceTag = "source"
	jsonTag   = "json"

	tagURL     = "url"
	tagBody    = "body"
	tagHeader  = "header"
	tagHeaders = "headers"
	tagCookie  = "cookie"
	tagCookies = "cookies"
)

// ParsedInput represents the parsed input data.
type ParsedInput struct {
	URLParameters map[string]any    // URL parameters for the request.
	Headers       map[string]string // HTTP headers to include in the request.
	Cookies       []http.Cookie     // Cookies to include in the request.
	Body          map[string]any    // Request body data.
}

// ParseInput parses the input struct and returns the parsed data. It will
// populate the URL parameters, headers, cookies, and body based on the
// struct tags. E.g. the struct field `source: url` will be placed in the URL.
//
//   - method: The HTTP method for the request.
//   - input: The input struct to parse.
func ParseInput(method string, input any) (*ParsedInput, error) {
	if input == nil {
		return nil, fmt.Errorf("parsed input is nil")
	}

	// Initialize maps and slices for parsed data
	headers := make(map[string]string)
	cookies := make([]http.Cookie, 0)
	urlParameters := make(map[string]any)
	body := make(map[string]any)

	// Extract values from the input struct and process them based on their tags
	inputVal := reflect.ValueOf(input).Elem()
	inputType := inputVal.Type()

	for i := 0; i < inputVal.NumField(); i++ {
		field := inputVal.Field(i)
		fieldInfo := inputType.Field(i)
		err := processField(
			field,
			fieldInfo,
			determineDefaultPlacement(method),
			headers,
			&cookies,
			urlParameters,
			body,
		)
		if err != nil {
			return nil, err
		}
	}

	return &ParsedInput{
		URLParameters: urlParameters,
		Headers:       headers,
		Cookies:       cookies,
		Body:          body,
	}, nil
}

// ConstructURL returns the full URL with query parameters.
//   - host: The host server to send the request to.
//   - path: The endpoint URL path.
//   - params: The query parameters as a string.
func ConstructURL(host string, path string, params string) string {
	url := fmt.Sprintf("%s%s", host, path)

	// Return the base URL if there are no URL parameters
	if len(params) == 0 {
		return url
	}

	return fmt.Sprintf("%s?%s", url, params)
}

// determineDefaultPlacement determines the default placement for a field based
// on the HTTP method.
func determineDefaultPlacement(method string) string {
	switch method {
	case http.MethodGet:
		return tagURL
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return tagBody
	default:
		return tagBody
	}
}

// processField processes a field of the input struct based on its tags.
func processField(
	field reflect.Value,
	fieldInfo reflect.StructField,
	defaultPlacement string,
	headers map[string]string,
	cookies *[]http.Cookie,
	urlParameters map[string]any,
	body map[string]any,
) error {
	// Determine field value placement in the request
	placement := fieldInfo.Tag.Get(sourceTag)
	if placement == "" {
		placement = defaultPlacement
	}

	// Place the field value in the appropriate map or slice
	return placeFieldValue(
		placement,
		determineFieldName(fieldInfo.Tag.Get(jsonTag), fieldInfo.Name),
		extractFieldValue(field),
		headers,
		cookies,
		urlParameters,
		body,
	)
}

// determineFieldName determines the field name to use in the request based on
// the JSON tag and the field name.
func determineFieldName(jsonTag string, fieldName string) string {
	jsonFieldName := strings.Split(jsonTag, ",")[0]
	if jsonFieldName == "" {
		jsonFieldName = fieldName
	}
	return jsonFieldName
}

// extractFieldValue extracts the field value based on its type.
func extractFieldValue(field reflect.Value) any {
	switch field.Kind() {
	case reflect.Bool:
		return field.Bool()
	case reflect.String:
		return field.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint()
	case reflect.Float32, reflect.Float64:
		return field.Float()
	default:
		return field.Interface()
	}
}

// placeFieldValue places the field value in the appropriate map or slice based
// on its placement.
func placeFieldValue(
	placement string,
	jsonFieldName string,
	value any,
	headers map[string]string,
	cookies *[]http.Cookie,
	urlParameters map[string]any,
	body map[string]any,
) error {
	switch placement {
	case tagURL:
		urlParameters[jsonFieldName] = value
	case tagBody:
		body[jsonFieldName] = value
	case tagHeader, tagHeaders:
		headers[jsonFieldName] = fmt.Sprintf("%v", value)
	case tagCookie, tagCookies:
		*cookies = append(*cookies, http.Cookie{
			Name:  jsonFieldName,
			Value: fmt.Sprintf("%v", value),
		})
	default:
		return fmt.Errorf("invalid source tag: %s", placement)
	}
	return nil
}
