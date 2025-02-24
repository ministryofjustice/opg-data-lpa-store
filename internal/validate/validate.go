package validate

import (
	"regexp"

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
