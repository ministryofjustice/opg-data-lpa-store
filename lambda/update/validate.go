package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

type Applyable interface {
	Apply(*shared.Lpa) []shared.FieldError
}

func validateUpdate(update shared.Update, lpa *shared.Lpa) (Applyable, []shared.FieldError) {
	switch update.Type {
	case "ATTORNEY_SIGN":
		return validateAttorneySign(update.Changes, lpa)
	case "CERTIFICATE_PROVIDER_OPT_OUT":
		return validateCertificateProviderOptOut(update.Changes)
	case "CERTIFICATE_PROVIDER_SIGN":
		return validateCertificateProviderSign(update.Changes, lpa)
	case "PERFECT":
		return validatePerfect(update.Changes)
	case "REGISTER":
		return validateRegister(update.Changes)
	case "TRUST_CORPORATION_SIGN":
		return validateTrustCorporationSign(update.Changes, lpa)
	case "DONOR_CONFIRM_IDENTITY":
		return validateDonorConfirmIdentity(update.Changes, lpa)
	case "CERTIFICATE_PROVIDER_CONFIRM_IDENTITY":
		return validateCertificateProviderConfirmIdentity(update.Changes, lpa)
	default:
		return nil, []shared.FieldError{{Source: "/type", Detail: "invalid value"}}
	}
}
