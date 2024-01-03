package main

import (
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
)

type CertificateProvider struct {
	UpdatedAt                 time.Time      `json:"updatedAt"`
	Address                   shared.Address `json:"address"`
	SignedAt                  time.Time      `json:"signedAt"`
	ContactLanguagePreference shared.Lang    `json:"contactLanguagePreference"`
}

func Validate(certificateProvider CertificateProvider) []shared.FieldError {
	return validate.All(
		validate.Address("/address", certificateProvider.Address),
		validate.Time("/signedAt", certificateProvider.SignedAt),
		validate.IsValid("/contactLanguagePreference", certificateProvider.ContactLanguagePreference),
	)
}
