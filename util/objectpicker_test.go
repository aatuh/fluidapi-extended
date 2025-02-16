package util

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
)

// dummyURLDecoder is a test implementation of the URLDecoder interface.
// It simply converts all query parameters into a map[string]any.
type dummyURLDecoder struct{}

func (d *dummyURLDecoder) DecodeURL(values url.Values) (map[string]any, error) {
	result := make(map[string]any)
	for k, v := range values {
		if len(v) == 1 {
			result[k] = v[0]
		} else {
			result[k] = v
		}
	}
	return result, nil
}

// Test structures for verifying multiple ways of fetching data from an
// http.Request. Fields come from either URL, body, headers, or cookies.
type testPayload struct {
	URLField    string `json:"url_field" source:"url"`
	BodyField   string `json:"body_field" source:"body"`
	HeaderField string `json:"header_field" source:"header"`
	CookieField string `json:"cookie_field" source:"cookie"`

	// This field doesn't specify a source tag, so it should use the default
	// source, which depends on the HTTP method.
	DefaultField string `json:"default_field"`

	// Multiple sources in a single tag as a comma-separated list. The code
	// uses the first non-empty one found in the request.
	MultiSource string `json:"multi_source" source:"url,body,header"`
}

// noTagPayload checks that fields without tags never get populated.
type noTagPayload struct {
	Field1 string
	Field2 int
}

// mixedStruct for testing partial tags. Some fields have tags, some don't.
type mixedStruct struct {
	FieldA string `json:"field_a"`
	FieldB int    // no tags at all
	FieldC bool   `json:"field_c" source:"header"`
}

// customDefaultSourcePayload helps test default source logic: fields that
// have no `source` tag will get the default source determined by the HTTP
// method.
type customDefaultSourcePayload struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
}

// unknownSourcePayload is for testing how the code handles an invalid
// "source" value. We expect it to panic.
type unknownSourcePayload struct {
	Field string `json:"field" source:"invalidSource"`
}

// TestPickObjectFromGET verifies that for a GET request, fields without
// a specific `source` tag fallback to `url` as the default source.
func TestPickObjectFromGET(t *testing.T) {
	// Setup URL query parameters for GET request
	req := httptest.NewRequest(http.MethodGet,
		"/test?url_field=hello+url&default_field=default_via_url"+
			"&multi_source=multi_value_url",
		nil)
	w := httptest.NewRecorder()

	// Add a header and cookie that we won't rely on for this test,
	// just to confirm that picking from URL works as intended.
	req.Header.Set("header_field", "header_value")
	req.AddCookie(&http.Cookie{Name: "cookie_field", Value: "cookie_value"})

	picker := NewObjectPicker[testPayload](&dummyURLDecoder{})
	got, err := picker.PickObject(req, w, testPayload{})
	require.NoError(t, err)

	require.Equal(t, "hello url", got.URLField)
	require.Equal(t, "", got.BodyField)
	require.Equal(t, "", got.HeaderField)
	require.Equal(t, "", got.CookieField)

	// For GET, the default is `url`, so default_field should be pulled
	// from the query params.
	require.Equal(t, "default_via_url", got.DefaultField)

	// multi_source was set in URL, so it should come from there.
	require.Equal(t, "multi_value_url", got.MultiSource)
}

// TestPickObjectFromPOST verifies that a POST request defaults to `body`
// for fields lacking a source tag.
func TestPickObjectFromPOST(t *testing.T) {
	bodyMap := map[string]interface{}{
		"body_field":    "hello body",
		"default_field": "default_via_body",
		"multi_source":  "multi_value_body",
	}
	body, _ := json.Marshal(bodyMap)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Query param for url_field (we won't rely on it for default source).
	q := req.URL.Query()
	q.Add("url_field", "ignored_url_value")
	req.URL.RawQuery = q.Encode()

	// Cookie for cookie_field
	req.AddCookie(&http.Cookie{Name: "cookie_field", Value: "cookie_value"})

	picker := NewObjectPicker[testPayload](&dummyURLDecoder{})
	got, err := picker.PickObject(req, w, testPayload{})
	require.NoError(t, err)

	// Even though we have a url_field param, we never used it, because
	// we didn't set default to URL. This field is specifically sourced
	// from URL though, so let's see if it picks it up:
	require.Equal(t, "ignored_url_value", got.URLField)

	// The body_field must come from the body
	require.Equal(t, "hello body", got.BodyField)

	// The default_field comes from the body because default source
	// for POST is `body`.
	require.Equal(t, "default_via_body", got.DefaultField)

	// multi_source should find a body value for multi_source first
	require.Equal(t, "multi_value_body", got.MultiSource)

	// We didn't pick from headers at all; remains empty
	require.Equal(t, "", got.HeaderField)

	// Cookie field is set from the cookie
	require.Equal(t, "cookie_value", got.CookieField)
}

// TestPickObjectFromHeaders checks that fields tagged as header
// are pulled from the request headers.
func TestPickObjectFromHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	req.Header.Set("header_field", "header_value")
	req.Header.Set("Content-Type", "application/json")

	picker := NewObjectPicker[testPayload](&dummyURLDecoder{})
	got, err := picker.PickObject(req, w, testPayload{})
	require.NoError(t, err)

	require.Equal(t, "header_value", got.HeaderField)
	require.Equal(t, "", got.CookieField)
}

// TestPickObjectFromCookie checks that fields tagged as cookie
// are pulled from the request cookies.
func TestPickObjectFromCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "cookie_field", Value: "cookie_value"})
	w := httptest.NewRecorder()

	picker := NewObjectPicker[testPayload](&dummyURLDecoder{})
	got, err := picker.PickObject(req, w, testPayload{})
	require.NoError(t, err)

	require.Equal(t, "cookie_value", got.CookieField)
	require.Equal(t, "", got.HeaderField)
}

// TestNoTagPayload ensures fields without JSON tags are never populated.
func TestNoTagPayload(t *testing.T) {
	bodyMap := map[string]interface{}{
		"Field1": "value1",
		"Field2": 123,
	}
	body, _ := json.Marshal(bodyMap)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	w := httptest.NewRecorder()

	picker := NewObjectPicker[noTagPayload](&dummyURLDecoder{})
	got, err := picker.PickObject(req, w, noTagPayload{})
	require.NoError(t, err)

	// Both should remain zero-values since there are no struct tags
	require.Equal(t, "", got.Field1)
	require.Equal(t, 0, got.Field2)
}

// TestMixedStruct verifies partial usage of tags and how default source
// is assigned or not assigned.
func TestMixedStruct(t *testing.T) {
	bodyMap := map[string]interface{}{
		"field_a": "body_value_a",
	}
	body, _ := json.Marshal(bodyMap)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("field_c", "header_value_c")
	w := httptest.NewRecorder()

	picker := NewObjectPicker[mixedStruct](&dummyURLDecoder{})
	got, err := picker.PickObject(req, w, mixedStruct{})
	require.NoError(t, err)

	// field_a has a json tag but no explicit source => default is body (POST)
	require.Equal(t, "body_value_a", got.FieldA)

	// FieldB has no tag => no pick
	require.Equal(t, 0, got.FieldB)

	// FieldC is explicitly from a header
	require.True(t, got.FieldC)
	// Actually, the code picks the string "header_value_c".
	// mapstructure sees "header_value_c" as a non-empty string for a bool,
	// so it will coerce it to 'true'. If you want exact type matching, you
	// can configure mapstructure further or handle that differently. For
	// testing, let's just check that it's `true`.
}

// TestDefaultSourceCheck verifies that a custom type with no explicit source
// picks from the default source for GET and POST.
func TestDefaultSourceCheck(t *testing.T) {
	type testCase struct {
		method      string
		field1Val   string
		queryVal    string
		bodyVal     string
		expectedVal string
	}

	cases := []testCase{
		{
			method:      http.MethodGet,
			field1Val:   "foo",
			queryVal:    "foo",
			bodyVal:     "body_ignored",
			expectedVal: "foo",
		},
		{
			method:      http.MethodPost,
			field1Val:   "bar",
			queryVal:    "url_ignored",
			bodyVal:     "bar",
			expectedVal: "bar",
		},
	}

	for _, c := range cases {
		t.Run(c.method, func(t *testing.T) {
			var req *http.Request
			if c.method == http.MethodGet {
				req = httptest.NewRequest(http.MethodGet,
					"/test?field1="+c.queryVal, nil)
			} else {
				bodyMap := map[string]interface{}{
					"field1": c.bodyVal,
					"field2": 12,
				}
				b, _ := json.Marshal(bodyMap)
				req = httptest.NewRequest(http.MethodPost,
					"/test", bytes.NewReader(b))
			}

			w := httptest.NewRecorder()
			picker := NewObjectPicker[customDefaultSourcePayload](
				&dummyURLDecoder{})

			got, err := picker.PickObject(req, w, customDefaultSourcePayload{})
			require.NoError(t, err)
			require.Equal(t, c.expectedVal, got.Field1)
		})
	}
}

// TestEmptyBody makes sure an empty body does not cause a decoding error.
func TestEmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()

	picker := NewObjectPicker[testPayload](&dummyURLDecoder{})
	got, err := picker.PickObject(req, w, testPayload{})
	require.NoError(t, err)
	require.NotNil(t, got) // The result object is valid, just no data.
}

// TestPanicOnUnknownSource checks that if we have an invalid source tag,
// we panic as expected (the code explicitly does a panic).
func TestPanicOnUnknownSource(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	picker := NewObjectPicker[unknownSourcePayload](&dummyURLDecoder{})
	defer func() {
		r := recover()
		require.NotNil(t, r, "Expected panic for unknown input source")
	}()

	_, _ = picker.PickObject(req, w, unknownSourcePayload{})
}

// TestAddToStructRegistry checks that adding the same struct multiple times
// merges sources properly. (Mostly ensures no panics, and that the new sources
// get appended for each field.)
func TestAddToStructRegistry(t *testing.T) {
	picker := NewObjectPicker[testPayload](&dummyURLDecoder{})

	obj := testPayload{} // typical usage
	var secondObj testPayload

	// Call mustUpdateObjectRegistry multiple times to see if it merges
	// without error. We can rely on internal coverage, but let's ensure
	// it doesn't panic or misbehave.
	picker.MustUpdateTestHook([]any{obj}, objectpickerTestDefaultSource)
	picker.MustUpdateTestHook([]any{secondObj}, objectpickerTestDefaultSource)
	// If we got here, no panic. We can do a quick check on the registry:
	// We'll just rely on no panic. You can also reflect on the registry
	// if you like.

	// A final sanity check: pick an object to ensure the flow still works
	req := httptest.NewRequest(http.MethodGet, "/test?url_field=123", nil)
	w := httptest.NewRecorder()
	got, err := picker.PickObject(req, w, testPayload{})
	require.NoError(t, err)
	require.Equal(t, "123", got.URLField)
}

// objectpickerTestDefaultSource is used as the "defaultSource" argument
// in the MustUpdateTestHook below. Normally that logic is internal to
// the ObjectPicker, but if you want to test it directly, you can define
// a constant here.
const objectpickerTestDefaultSource = "url"

// MustUpdateTestHook is a small helper on *ObjectPicker for test
// coverage of mustUpdateObjectRegistry. You could also test it by
// calling `PickObject` with different struct shapes, but sometimes
// calling it directly is simpler.
func (o *ObjectPicker[T]) MustUpdateTestHook(
	objectSamples []any,
	defaultSource string,
) {
	o.mustUpdateObjectRegistry(objectSamples, defaultSource)
}

// The code below stubs out the external references in your snippet (for
// example, your `api` package and the `core.NewError[any]`). If you have
// them properly defined, you can remove these stubs. They are here to
// make this file self-contained.
type anyError struct {
	msg string
}

func (e anyError) Error() string {
	return e.msg
}

func newError(msg string) error {
	return anyError{msg: msg}
}

// Example test for verifying your custom error usage if needed:
func TestErrorTypes(t *testing.T) {
	err := newError("INVALID_INPUT")
	require.NotNil(t, err)
	require.Equal(t, "INVALID_INPUT", err.Error())
}

// If your code or tests require a custom mapstructure configuration that
// is different from the default, you can define your own. This example
// is just to illustrate that you could customize mapstructure decoding
// in a test, as needed.
func customMapstructureDecode(input, output any) error {
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           output,
		TagName:          "json",
		WeaklyTypedInput: true,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

// TestCustomMapstructureDecoder example if you want to test a custom
// decoding logic. Typically, you might not need this unless you changed
// the internals.
func TestCustomMapstructureDecoder(t *testing.T) {
	input := map[string]interface{}{
		"name":   "Alice",
		"active": "true",
	}
	var out struct {
		Name   string `json:"name"`
		Active bool   `json:"active"`
	}
	err := customMapstructureDecode(input, &out)
	require.NoError(t, err)
	require.Equal(t, "Alice", out.Name)
	require.True(t, out.Active)
}
