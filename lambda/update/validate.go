package main

import (
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
)

var (
	detailMustBeString = "must be a string"
)

type CertificateProviderSign struct {
	Address                   shared.Address
	SignedAt                  time.Time
	ContactLanguagePreference shared.Lang
}

func validateUpdate(update shared.Update) []shared.FieldError {
	switch update.Type {
	case "CERTIFICATE_PROVIDER_SIGN":
		var errors []shared.FieldError
		var ok bool
		x := CertificateProviderSign{}

		for i, change := range update.Changes {
			if change.Old != nil {
				errors = append(errors, shared.FieldError{Source: fmt.Sprintf("/changes/%d/old", i), Detail: "field must not be provided"})
			}

			newKey := fmt.Sprintf("/changes/%d/new", i)
			switch change.Key {
			case "/certificateProvider/address/line1":
				if x.Address.Line1, ok = change.New.(string); !ok {
					errors = append(errors, shared.FieldError{Source: newKey, Detail: detailMustBeString})
				}
			case "/certificateProvider/address/line2":
				if x.Address.Line2, ok = change.New.(string); !ok {
					errors = append(errors, shared.FieldError{Source: newKey, Detail: detailMustBeString})
				}
			case "/certificateProvider/address/line3":
				if x.Address.Line3, ok = change.New.(string); !ok {
					errors = append(errors, shared.FieldError{Source: newKey, Detail: detailMustBeString})
				}
			case "/certificateProvider/address/town":
				if x.Address.Town, ok = change.New.(string); !ok {
					errors = append(errors, shared.FieldError{Source: newKey, Detail: detailMustBeString})
				}
			case "/certificateProvider/address/postcode":
				if x.Address.Postcode, ok = change.New.(string); !ok {
					errors = append(errors, shared.FieldError{Source: newKey, Detail: detailMustBeString})
				}
			case "/certificateProvider/address/country":
				if x.Address.Country, ok = change.New.(string); !ok {
					errors = append(errors, shared.FieldError{Source: newKey, Detail: detailMustBeString})
				} else {
					errors = append(errors, validate.Country(newKey, x.Address.Country)...)
				}
			case "/certificateProvider/signedAt":
				if x.SignedAt, ok = change.New.(time.Time); !ok {
					errors = append(errors, shared.FieldError{Source: newKey, Detail: "must be a datetime"})
				}
			case "/certificateProvider/contactLanguagePreference":
				if x.ContactLanguagePreference, ok = change.New.(shared.Lang); !ok {
					errors = append(errors, shared.FieldError{Source: newKey, Detail: detailMustBeString})
				}
			default:
				errors = append(errors, shared.FieldError{Source: fmt.Sprintf("/changes/%d", i), Detail: "change not allowed for type"})
			}
		}

		if x.Address.IsSet() {
			if x.Address.Line1 == "" {
				errors = append(errors, shared.FieldError{Source: "/changes", Detail: "missing /certificateProvider/address/line1"})
			}

			if x.Address.Town == "" {
				errors = append(errors, shared.FieldError{Source: "/changes", Detail: "missing /certificateProvider/address/town"})
			}

			if x.Address.Country == "" {
				errors = append(errors, shared.FieldError{Source: "/changes", Detail: "missing /certificateProvider/address/country"})
			}
		}

		if x.SignedAt.IsZero() {
			errors = append(errors, shared.FieldError{Source: "/changes", Detail: "missing /certificateProvider/signedAt"})
		}

		if x.ContactLanguagePreference == shared.Lang("") {
			errors = append(errors, shared.FieldError{Source: "/changes", Detail: "missing /certificateProvider/contactLanguagePreference"})
		}

		return errors

	default:
		return []shared.FieldError{}
	}
}
