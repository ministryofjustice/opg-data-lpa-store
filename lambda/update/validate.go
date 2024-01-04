package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

type Applyable interface {
	Apply(*shared.Lpa) []shared.FieldError
}

func validateUpdate(update shared.Update) (Applyable, []shared.FieldError) {
	switch update.Type {
	case "CERTIFICATE_PROVIDER_SIGN":
		return validateCertificateProviderSign(update.Changes)
	case "ATTORNEY_SIGN":
		return validateAttorneySign(update.Changes)
	case "TRUST_CORPORATION_SIGN":
		return validateTrustCorporationSign(update.Changes)
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
