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
	case "PERFECT", "STATUTORY_WAITING_PERIOD":
		return validateStatutoryWaitingPeriod(update.Changes)
	case "REGISTER":
		return validateRegister(update.Changes)
	case "OPG_CHANGE_STATUS":
		return validateOpgChangeStatus(update.Changes)
	case "TRUST_CORPORATION_SIGN":
		return validateTrustCorporationSign(update.Changes, lpa)
	case "DONOR_CONFIRM_IDENTITY":
		return validateDonorConfirmIdentity(update.Changes, lpa)
	case "CERTIFICATE_PROVIDER_CONFIRM_IDENTITY":
		return validateCertificateProviderConfirmIdentity(update.Changes, lpa)
	case "DONOR_WITHDRAW_LPA":
		return validateDonorWithdrawLPA(update.Changes)
	case "ATTORNEY_OPT_OUT":
		return validateAttorneyOptOut(update)
	case "TRUST_CORPORATION_OPT_OUT":
		return validateTrustCorporationOptOut(update)
	default:
		return nil, []shared.FieldError{{Source: "/type", Detail: "invalid value"}}
	}
}
