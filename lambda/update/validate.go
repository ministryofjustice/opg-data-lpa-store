package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
)

type Applyable interface {
	Apply(*shared.Lpa) []shared.FieldError
}

type CertificateProviderSign struct {
	Address                   shared.Address
	SignedAt                  time.Time
	ContactLanguagePreference shared.Lang
}

func (c CertificateProviderSign) Apply(lpa *shared.Lpa) []shared.FieldError {
	lpa.CertificateProvider.Address = c.Address
	lpa.CertificateProvider.SignedAt = c.SignedAt
	lpa.CertificateProvider.ContactLanguagePreference = c.ContactLanguagePreference

	return nil
}

func validateUpdate(update shared.Update) (Applyable, []shared.FieldError) {
	switch update.Type {
	case "CERTIFICATE_PROVIDER_SIGN":
		var (
			data   CertificateProviderSign
			errors []shared.FieldError
		)

		for i, change := range update.Changes {
			if len(change.Old) != 0 {
				errors = append(errors, shared.FieldError{Source: fmt.Sprintf("/changes/%d/old", i), Detail: "field must not be provided"})
			}

			newKey := fmt.Sprintf("/changes/%d/new", i)
			switch change.Key {
			case "/certificateProvider/address/line1":
				if err := json.Unmarshal(change.New, &data.Address.Line1); err != nil {
					errors = errorMustBeString(errors, newKey)
				}
			case "/certificateProvider/address/line2":
				if err := json.Unmarshal(change.New, &data.Address.Line2); err != nil {
					errors = errorMustBeString(errors, newKey)
				}
			case "/certificateProvider/address/line3":
				if err := json.Unmarshal(change.New, &data.Address.Line3); err != nil {
					errors = errorMustBeString(errors, newKey)
				}
			case "/certificateProvider/address/town":
				if err := json.Unmarshal(change.New, &data.Address.Town); err != nil {
					errors = errorMustBeString(errors, newKey)
				}
			case "/certificateProvider/address/postcode":
				if err := json.Unmarshal(change.New, &data.Address.Postcode); err != nil {
					errors = errorMustBeString(errors, newKey)
				}
			case "/certificateProvider/address/country":
				if err := json.Unmarshal(change.New, &data.Address.Country); err != nil {
					errors = errorMustBeString(errors, newKey)
				} else {
					errors = append(errors, validate.Country(newKey, data.Address.Country)...)
				}
			case "/certificateProvider/signedAt":
				if err := json.Unmarshal(change.New, &data.SignedAt); err != nil {
					errors = errorMustBeDateTime(errors, newKey)
				}
			case "/certificateProvider/contactLanguagePreference":
				if err := json.Unmarshal(change.New, &data.ContactLanguagePreference); err != nil {
					errors = errorMustBeString(errors, newKey)
				} else {
					errors = append(errors, validate.IsValid(newKey, data.ContactLanguagePreference)...)
				}
			default:
				errors = append(errors, shared.FieldError{Source: fmt.Sprintf("/changes/%d", i), Detail: "change not allowed for type"})
			}
		}

		if data.Address.IsSet() {
			if data.Address.Line1 == "" {
				errors = errorMissing(errors, "/certificateProvider/address/line1")
			}

			if data.Address.Town == "" {
				errors = errorMissing(errors, "/certificateProvider/address/town")
			}

			if data.Address.Country == "" {
				errors = errorMissing(errors, "/certificateProvider/address/country")
			}
		}

		if data.SignedAt.IsZero() {
			errors = errorMissing(errors, "/certificateProvider/signedAt")
		}

		if data.ContactLanguagePreference == shared.Lang("") {
			errors = errorMissing(errors, "/certificateProvider/contactLanguagePreference")
		}

		return data, errors

	default:
		return nil, []shared.FieldError{{Source: "/type", Detail: "invalid value"}}
	}
}

func errorMustBeString(errors []shared.FieldError, source string) []shared.FieldError {
	return append(errors, shared.FieldError{Source: source, Detail: "must be a string"})
}

func errorMustBeDateTime(errors []shared.FieldError, source string) []shared.FieldError {
	return append(errors, shared.FieldError{Source: source, Detail: "must be a datetime"})
}

func errorMissing(errors []shared.FieldError, key string) []shared.FieldError {
	return append(errors, shared.FieldError{Source: "/changes", Detail: "missing " + key})
}
