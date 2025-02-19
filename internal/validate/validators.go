package validate

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

const (
	msgRequired    = "field is required"
	msgType        = "unexpected type"
	msgNotProvided = "field must not be provided"
	msgInvalid     = "invalid value"
	msgFormat      = "invalid format"
	msgCountryCode = "must be a valid ISO-3166-1 country code"
	msgDateTime    = "must be a valid datetime"
)

type Validator interface {
	Valid(val any) string
}

type NotEmptyValidator struct{}

func (v NotEmptyValidator) Valid(val any) string {
	switch v := val.(type) {
	case *string:
		if *v == "" {
			return msgRequired
		} else {
			return ""
		}
	case string:
		if v == "" {
			return msgRequired
		} else {
			return ""
		}
	case *time.Time:
		if v == nil || v.IsZero() {
			return msgRequired
		} else {
			return ""
		}
	case time.Time:
		if v.IsZero() {
			return msgRequired
		} else {
			return ""
		}
	}

	return msgType
}

func NotEmpty() Validator {
	return NotEmptyValidator{}
}

type EmptyValidator struct{}

func (v EmptyValidator) Valid(val any) string {
	switch v := val.(type) {
	case *string:
		if *v == "" {
			return ""
		} else {
			return msgNotProvided
		}
	case string:
		if v == "" {
			return ""
		} else {
			return msgNotProvided
		}
	case *time.Time:
		if v == nil || v.IsZero() {
			return ""
		} else {
			return msgNotProvided
		}
	case time.Time:
		if v.IsZero() {
			return ""
		} else {
			return msgNotProvided
		}
	}

	return msgType
}

func Empty() Validator {
	return EmptyValidator{}
}

type ValidValidator struct{}

func (v ValidValidator) Valid(val any) string {
	if fmt.Sprint(val) == "" {
		return msgRequired
	}

	enum, ok := val.(interface{ IsValid() bool })
	if !ok {
		return msgType
	}

	if !enum.IsValid() {
		return msgInvalid
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
		return msgType
	}

	if !unset.Unset() {
		return msgNotProvided
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
		return msgType
	}

	if str == "" {
		return msgRequired
	}

	if uuid.Validate(str) != nil {
		return msgFormat
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
			return msgFormat
		}

		if v.IsZero() {
			return msgRequired
		}

		return ""
	case *shared.Date:
		if v.IsMalformed {
			return msgFormat
		}

		if v.IsZero() {
			return msgRequired
		}

		return ""
	}

	return msgType
}

func Date() Validator {
	return DateValidator{}
}

type CountryValidator struct{}

func (v CountryValidator) Valid(val any) string {
	switch v := val.(type) {
	case string:
		if !countryCodeRe.MatchString(v) {
			return msgCountryCode
		} else {
			return ""
		}
	case *string:
		if !countryCodeRe.MatchString(*v) {
			return msgCountryCode
		} else {
			return ""
		}
	}

	return msgType
}

func Country() Validator {
	return CountryValidator{}
}

type OptionalTimeValidator struct{}

func (v OptionalTimeValidator) Valid(val any) string {
	t, ok := val.(*time.Time)
	if !ok {
		return msgType
	}

	if t != nil && t.IsZero() {
		return msgDateTime
	}

	return ""
}

func OptionalTime() Validator {
	return OptionalTimeValidator{}
}
