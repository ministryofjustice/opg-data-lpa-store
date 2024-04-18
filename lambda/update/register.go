package main

import (
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

type Register struct{}

func (r Register) Apply(lpa *shared.Lpa) []shared.FieldError {
	if lpa.RegistrationDate != nil && !lpa.RegistrationDate.IsZero() {
		return []shared.FieldError{{Source: "/type", Detail: "lpa already registered"}}
	}

	if lpa.SignedAt.IsZero() {
		return []shared.FieldError{{Source: "/type", Detail: "lpa must be signed"}}
	}

	if lpa.CertificateProvider.SignedAt == nil || lpa.CertificateProvider.SignedAt.IsZero() {
		return []shared.FieldError{{Source: "/type", Detail: "lpa must have a certificate"}}
	}

	for _, a := range lpa.Attorneys {
		if a.SignedAt == nil || a.SignedAt.IsZero() {
			return []shared.FieldError{{Source: "/type", Detail: "lpa must be signed by attorneys"}}
		}
	}

	for _, t := range lpa.TrustCorporations {
		for _, s := range t.Signatories {
			if s.SignedAt.IsZero() {
				return []shared.FieldError{{Source: "/type", Detail: "lpa must be signed by trust corporations"}}
			}
		}
	}

	now := time.Now().UTC()
	lpa.RegistrationDate = &now
	lpa.Status = shared.LpaStatusRegistered

	return nil
}

func validateRegister(changes []shared.Change) (Register, []shared.FieldError) {
	if len(changes) > 0 {
		return Register{}, []shared.FieldError{{Source: "/changes", Detail: "expected empty"}}
	}

	return Register{}, nil
}
