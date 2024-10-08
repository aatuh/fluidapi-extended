package urlencoder

import (
	"net/url"
	"reflect"
)

type URLEncoder struct{}

func (e URLEncoder) EncodeURL(data map[string]any) (url.Values, error) {
	values := url.Values{}
	for key, value := range data {
		err := EncodeURL(&values, key, reflect.ValueOf(value))
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}

func (e URLEncoder) DecodeURL(values url.Values) (map[string]any, error) {
	return DecodeURL(values)
}
