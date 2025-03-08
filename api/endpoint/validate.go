package endpoint

import (
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
)

type Validate struct {
	customRules map[string]func(any) error
}

func NewValidate(customRules map[string]func(any) error) *Validate {
	return &Validate{
		customRules: customRules,
	}
}

// MustFromRules is like FromRules but panics on error.
func (v *Validate) MustFromRules(rules []string) func(any) error {
	val, err := v.FromRules(rules)
	if err != nil {
		panic(err)
	}
	return val
}

// FromRules creates a validator function from a list of rules.
func (v *Validate) FromRules(rules []string) (func(any) error, error) {
	vType := strings.ToLower(rules[0])
	if v.customRules[vType] != nil {
		return v.customRules[vType], nil
	}
	switch vType {
	case "string":
		var validators []StringValidator
		for _, rule := range rules[1:] {
			parts := strings.SplitN(rule, "=", 2)
			key := strings.ToLower(parts[0])
			var param string
			if len(parts) == 2 {
				param = parts[1]
			}

			switch key {
			// case "required":
			// 	validators = append(validators, v.RequiredString())
			case "len":
				n, err := strconv.Atoi(param)
				if err != nil {
					return nil, fmt.Errorf("invalid parameter for len: %w", err)
				}
				validators = append(validators, v.Length(n))
			case "min":
				n, err := strconv.Atoi(param)
				if err != nil {
					return nil, fmt.Errorf("invalid parameter for min: %w", err)
				}
				validators = append(validators, v.MinLength(n))
			case "max":
				n, err := strconv.Atoi(param)
				if err != nil {
					return nil, fmt.Errorf("invalid parameter for max: %w", err)
				}
				validators = append(validators, v.MaxLength(n))
			case "oneof":
				opts := strings.Split(param, " ")
				validators = append(validators, v.OneOf(opts...))
			case "email":
				validators = append(validators, v.Email())
			default:
				return nil, fmt.Errorf("unknown string validator: %s", key)
			}
		}
		return v.WithString(validators...), nil

	case "int":
		var validators []IntValidator
		for _, rule := range rules[1:] {
			parts := strings.SplitN(rule, "=", 2)
			key := strings.ToLower(parts[0])
			var param string
			if len(parts) == 2 {
				param = parts[1]
			}

			switch key {
			// case "required":
			// 	validators = append(validators, v.RequiredInt())
			case "min":
				n, err := strconv.ParseInt(param, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid parameter for min: %w", err)
				}
				validators = append(validators, v.MinInt(n))
			case "max":
				n, err := strconv.ParseInt(param, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid parameter for max: %w", err)
				}
				validators = append(validators, v.MaxInt(n))
			default:
				return nil, fmt.Errorf("unknown int validator: %s", key)
			}
		}
		return v.WithInt(validators...), nil

	case "int64":
		// This branch validates explicit int64 types.
		var validators []IntValidator
		for _, rule := range rules[1:] {
			parts := strings.SplitN(rule, "=", 2)
			key := strings.ToLower(parts[0])
			var param string
			if len(parts) == 2 {
				param = parts[1]
			}

			switch key {
			// case "required":
			// 	validators = append(validators, v.RequiredInt())
			case "min":
				n, err := strconv.ParseInt(param, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid parameter for min: %w", err)
				}
				validators = append(validators, v.MinInt(n))
			case "max":
				n, err := strconv.ParseInt(param, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid parameter for max: %w", err)
				}
				validators = append(validators, v.MaxInt(n))
			default:
				return nil, fmt.Errorf("unknown int64 validator: %s", key)
			}
		}
		return v.WithExplicitInt(validators...), nil
	case "bool":
		return v.WithBool(), nil
	case "slice":
		var validators []SliceValidator
		for _, rule := range rules[1:] {
			parts := strings.SplitN(rule, "=", 2)
			key := strings.ToLower(parts[0])
			var param string
			if len(parts) == 2 {
				param = parts[1]
			}

			switch key {
			case "len":
				n, err := strconv.ParseInt(param, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid parameter for len: %w", err)
				}
				validators = append(validators, v.SliceLength(int(n)))
			case "min":
				n, err := strconv.ParseInt(param, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid parameter for minlen: %w", err)
				}
				validators = append(validators, v.MinSliceLength(int(n)))
			case "max":
				n, err := strconv.ParseInt(param, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid parameter for maxlen: %w", err)
				}
				validators = append(validators, v.MaxSliceLength(int(n)))
			default:
				return nil, fmt.Errorf("unknown slice validator: %s", key)
			}
		}
		return v.WithSlice(validators...), nil
	default:
		return nil, fmt.Errorf("unknown validator type: %s", vType)
	}
}

// toBool converts a value to a bool.
func (v *Validate) toBool(value any) (bool, error) {
	b, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("value is not a bool")
	}
	return b, nil
}

// StringValidator is a function that validates a string.
type StringValidator func(s string) error

// WithString accepts a list of StringValidators and returns a function
// that accepts any type, converts it to a string, and then pipes the string
// through each validator in order.
func (v *Validate) WithString(validators ...StringValidator) func(value any) error {
	return func(value any) error {
		s, err := v.toString(value)
		if err != nil {
			return err
		}
		for _, validator := range validators {
			if err := validator(s); err != nil {
				return err
			}
		}
		return nil
	}
}

// RequiredString ensures the string is not empty.
func (v *Validate) RequiredString() StringValidator {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return errors.New("value is required")
		}
		return nil
	}
}

// Length returns a validator that ensures the string is exactly n characters long.
func (v *Validate) Length(n int) StringValidator {
	return func(s string) error {
		if len(s) != n {
			return fmt.Errorf("must be exactly %d characters long", n)
		}
		return nil
	}
}

// MinLength returns a validator that ensures the string is at least n characters long.
func (v *Validate) MinLength(n int) StringValidator {
	return func(s string) error {
		if len(s) < n {
			return fmt.Errorf("must be at least %d characters long", n)
		}
		return nil
	}
}

// MaxLength returns a validator that ensures the string is at most n characters long.
func (v *Validate) MaxLength(n int) StringValidator {
	return func(s string) error {
		if len(s) > n {
			return fmt.Errorf("must be at most %d characters long", n)
		}
		return nil
	}
}

// OneOf returns a validator that ensures the string is one of the given values.
func (v *Validate) OneOf(values ...string) StringValidator {
	return func(s string) error {
		for _, val := range values {
			if strings.EqualFold(s, val) {
				return nil
			}
		}
		return fmt.Errorf("must be one of %s", strings.Join(values, ", "))
	}
}

// Email returns a validator that ensures the string is a valid email address.
func (v *Validate) Email() StringValidator {
	return func(s string) error {
		if len(s) > 254 {
			return errors.New("must be at most 254 characters long")
		}
		if _, err := mail.ParseAddress(s); err != nil {
			return errors.New("must be a valid email address")
		}
		return nil
	}
}

// toString tries to convert any value to a string.
// If the value is already a string it returns it directly.
// If the value implements fmt.Stringer, its String method is used.
// Otherwise it returns an error.
func (v *Validate) toString(value any) (string, error) {
	if s, ok := value.(string); ok {
		return s, nil
	}
	if stringer, ok := value.(fmt.Stringer); ok {
		return stringer.String(), nil
	}
	return "", errors.New("cannot convert value to string")
}

// IntValidator is a function that validates an int64.
type IntValidator func(i int64) error

// WithInt takes many IntValidators and returns a single one
// that runs them in sequence.
func (v *Validate) WithInt(validators ...IntValidator) func(value any) error {
	return func(value any) error {
		i, err := v.toInt64(value)
		if err != nil {
			return err
		}
		for _, validator := range validators {
			if err := validator(i); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithExplicitInt is similar to WithInt but it only accepts values of type
// int64.
func (v *Validate) WithExplicitInt(validators ...IntValidator) func(value any) error {
	return func(value any) error {
		i, err := v.toExplicitInt64(value)
		if err != nil {
			return err
		}
		for _, validator := range validators {
			if err := validator(i); err != nil {
				return err
			}
		}
		return nil
	}
}

// RequiredInt returns a validator that ensures an integer is not 0.
func (v *Validate) RequiredInt() IntValidator {
	return func(i int64) error {
		if i == 0 {
			return errors.New("value is required")
		}
		return nil
	}
}

// MinInt returns a validator that ensures an integer is at least min.
func (v *Validate) MinInt(min int64) IntValidator {
	return func(i int64) error {
		if i < min {
			return fmt.Errorf("must be at least %d", min)
		}
		return nil
	}
}

// MaxInt returns a validator that ensures an integer is at most max.
func (v *Validate) MaxInt(max int64) IntValidator {
	return func(i int64) error {
		if i > max {
			return fmt.Errorf("must be at most %d", max)
		}
		return nil
	}
}

// toInt64 converts common integer types to int64.
func (v *Validate) toInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	default:
		return 0, errors.New("value is not an integer")
	}
}

// toExplicitInt64 asserts that the value is exactly an int64.
func (v *Validate) toExplicitInt64(value any) (int64, error) {
	if v, ok := value.(int64); ok {
		return v, nil
	}
	return 0, errors.New("value is not an int64")
}

// BoolValidator is a function that validates a bool.
type BoolValidator func(b bool) error

// WithBool is a function that validates a bool.
func (v *Validate) WithBool(validators ...BoolValidator) func(value any) error {
	return func(value any) error {
		_, err := v.toBool(value)
		if err != nil {
			return err
		}
		return nil
	}
}

// SliceValidator is a function that validates a slice.
type SliceValidator func(s []any) error

// WithSlice is a function that validates a slice.
func (v *Validate) WithSlice(validators ...SliceValidator) func(value any) error {
	return func(value any) error {
		s, err := v.toSlice(value)
		if err != nil {
			return err
		}
		for _, validator := range validators {
			if err := validator(s); err != nil {
				return err
			}
		}
		return nil
	}
}

// SliceLength returns a validator that ensures the slice is exactly n elements long.
func (v *Validate) SliceLength(n int) SliceValidator {
	return func(s []any) error {
		if len(s) != n {
			return fmt.Errorf("must have %d elements", n)
		}
		return nil
	}
}

// MinSliceLength returns a validator that ensures the slice is at least n elements long.
func (v *Validate) MinSliceLength(n int) SliceValidator {
	return func(s []any) error {
		if len(s) < n {
			return fmt.Errorf("must have at least %d elements", n)
		}
		return nil
	}
}

// MaxSliceLength returns a validator that ensures the slice is at most n elements long.
func (v *Validate) MaxSliceLength(n int) SliceValidator {
	return func(s []any) error {
		if len(s) > n {
			return fmt.Errorf("must have at most %d elements", n)
		}
		return nil
	}
}

// toSlice converts a value to a slice.
func (v *Validate) toSlice(value any) ([]any, error) {
	if s, ok := value.([]any); ok {
		return s, nil
	}
	return nil, errors.New("value is not a slice")
}
