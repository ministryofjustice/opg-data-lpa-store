package validate

import (
	"fmt"
	"regexp"
	"time"

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

func Required(source string, value string) []shared.FieldError {
	return If(value == "", []shared.FieldError{{Source: source, Detail: "field is required"}})
}

func Empty(source string, value string) []shared.FieldError {
	return If(value != "", []shared.FieldError{{Source: source, Detail: "field must not be provided"}})
}

func Date(source string, date shared.Date) []shared.FieldError {
	if date.IsMalformed {
		return []shared.FieldError{{Source: source, Detail: "invalid format"}}
	}

	if date.IsZero() {
		return []shared.FieldError{{Source: source, Detail: "field is required"}}
	}

	return nil
}

func Time(source string, t time.Time) []shared.FieldError {
	return If(t.IsZero(), []shared.FieldError{{Source: source, Detail: "field is required"}})
}

func Address(prefix string, address shared.Address) []shared.FieldError {
	return All(
		Required(fmt.Sprintf("%s/line1", prefix), address.Line1),
		Required(fmt.Sprintf("%s/town", prefix), address.Town),
		Required(fmt.Sprintf("%s/country", prefix), address.Country),
		If(!countryCodeRe.MatchString(address.Country), []shared.FieldError{{Source: fmt.Sprintf("%s/country", prefix), Detail: "must be a valid ISO-3166-1 country code"}}),
	)
}

type isValid interface {
	~string
	IsValid() bool
}

func IsValid[V isValid](source string, v V) []shared.FieldError {
	if e := Required(source, string(v)); e != nil {
		return e
	}

	if !v.IsValid() {
		return []shared.FieldError{{Source: source, Detail: "invalid value"}}
	}

	return nil
}

func Unset(source string, v interface{ Unset() bool }) []shared.FieldError {
	return If(!v.Unset(), []shared.FieldError{{Source: source, Detail: "field must not be provided"}})
}
