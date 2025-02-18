package validate

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

var countryCodeRe = regexp.MustCompile("^[A-Z]{2}$")

func All(fieldErrors ...[]shared.FieldError) []shared.FieldError {
	var errors []shared.FieldError

	for _, e := range fieldErrors {
		if e != nil {
			errors = append(errors, e...)
		}
	}

	return errors
}

func IfElse(ok bool, eIf []shared.FieldError, eElse []shared.FieldError) []shared.FieldError {
	if ok {
		return eIf
	}

	return eElse
}

func If(ok bool, e []shared.FieldError) []shared.FieldError {
	return IfElse(ok, e, nil)
}

func IfFunc(ok bool, fn func() []shared.FieldError) []shared.FieldError {
	if ok {
		return fn()
	}

	return nil
}

func WithSource(source string, val any, validators ...Validator) (errs []shared.FieldError) {
	for _, validator := range validators {
		msg := validator.Valid(val)
		if msg != "" {
			errs = append(errs, shared.FieldError{Source: source, Detail: msg})
		}
	}

	return errs
}

type Validator interface {
	Valid(val any) string
}

type NotEmptyValidator struct{}

func (v NotEmptyValidator) Valid(val any) string {
	const msg = "field is required"

	switch v := val.(type) {
	case *string:
		if *v == "" {
			return msg
		} else {
			return ""
		}
	case string:
		if v == "" {
			return msg
		} else {
			return ""
		}
	case *time.Time:
		if v == nil || v.IsZero() {
			return msg
		} else {
			return ""
		}
	case time.Time:
		if v.IsZero() {
			return msg
		} else {
			return ""
		}
	}

	return "unexpected type"
}

func NotEmpty() Validator {
	return NotEmptyValidator{}
}

type EmptyValidator struct{}

func (v EmptyValidator) Valid(val any) string {
	const msg = "field must not be provided"

	switch v := val.(type) {
	case *string:
		if *v == "" {
			return ""
		} else {
			return msg
		}
	case string:
		if v == "" {
			return ""
		} else {
			return msg
		}
	case *time.Time:
		if v == nil || v.IsZero() {
			return ""
		} else {
			return msg
		}
	case time.Time:
		if v.IsZero() {
			return ""
		} else {
			return msg
		}
	}

	return "unexpected type"
}

func Empty() Validator {
	return EmptyValidator{}
}

type ValidValidator struct{}

func (v ValidValidator) Valid(val any) string {
	if fmt.Sprint(val) == "" {
		return "field is required"
	}

	enum, ok := val.(interface{ IsValid() bool })
	if !ok {
		return "unexpected type"
	}

	if !enum.IsValid() {
		return "invalid value"
	}

	return ""
}

func Valid() Validator {
	return ValidValidator{}
}

type UnsetValidator struct{}

func (v UnsetValidator) Valid(val any) string {
	unset, ok := val.(interface{ Unset() bool })
	if !ok {
		return "unexpected type"
	}

	if !unset.Unset() {
		return "field must not be provided"
	}

	return ""
}

func Unset() Validator {
	return UnsetValidator{}
}

type UUIDValidator struct{}

func (v UUIDValidator) Valid(val any) string {
	str, ok := val.(string)
	if !ok {
		return "unexpected type"
	}

	if str == "" {
		return "field is required"
	}

	if uuid.Validate(str) != nil {
		return "invalid format"
	}

	return ""
}

func UUID() Validator {
	return UUIDValidator{}
}

type DateValidator struct{}

func (v DateValidator) Valid(val any) string {
	switch v := val.(type) {
	case shared.Date:
		if v.IsMalformed {
			return "invalid format"
		}

		if v.IsZero() {
			return "field is required"
		}

		return ""
	case *shared.Date:
		if v.IsMalformed {
			return "invalid format"
		}

		if v.IsZero() {
			return "field is required"
		}

		return ""
	}

	return "unexpected type"
}

func Date() Validator {
	return DateValidator{}
}

type CountryValidator struct{}

func (v CountryValidator) Valid(val any) string {
	switch v := val.(type) {
	case string:
		if !countryCodeRe.MatchString(v) {
			return "must be a valid ISO-3166-1 country code"
		} else {
			return ""
		}
	case *string:
		if !countryCodeRe.MatchString(*v) {
			return "must be a valid ISO-3166-1 country code"
		} else {
			return ""
		}
	}

	return "unexpected type"
}

func Country() Validator {
	return CountryValidator{}
}

type OptionalTimeValidator struct{}

func (v OptionalTimeValidator) Valid(val any) string {
	t, ok := val.(*time.Time)
	if !ok {
		return "unexpected type"
	}

	if t != nil && t.IsZero() {
		return "must be a valid datetime"
	}

	return ""
}

func OptionalTime() Validator {
	return OptionalTimeValidator{}
}
