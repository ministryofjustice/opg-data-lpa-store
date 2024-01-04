package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
)

type CertificateProviderSign struct {
	Address                   shared.Address
	SignedAt                  time.Time
	ContactLanguagePreference shared.Lang
}

func (c CertificateProviderSign) Apply(lpa *shared.Lpa) []shared.FieldError {
	if !lpa.CertificateProvider.SignedAt.IsZero() {
		return []shared.FieldError{{Source: "/type", Detail: "certificate provider cannot sign again"}}
	}

	lpa.CertificateProvider.Address = c.Address
	lpa.CertificateProvider.SignedAt = c.SignedAt
	lpa.CertificateProvider.ContactLanguagePreference = c.ContactLanguagePreference

	return nil
}

func validateCertificateProviderSign(changes []shared.Change) (data CertificateProviderSign, errors []shared.FieldError) {
	for i, change := range changes {
		if !bytes.Equal(change.Old, []byte("null")) {
			errors = append(errors, shared.FieldError{Source: fmt.Sprintf("/changes/%d/old", i), Detail: "field must be null"})
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

	if !data.Address.IsZero() {
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
}
