package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

type StatutoryWaitingPeriod struct{}

func (r StatutoryWaitingPeriod) Apply(lpa *shared.Lpa) []shared.FieldError {
	if lpa.Status != shared.LpaStatusInProgress {
		return []shared.FieldError{{Source: "/type", Detail: "status must be in-progress to enter statutory-waiting-period"}}
	}

	if lpa.SignedAt.IsZero() {
		return []shared.FieldError{{Source: "/type", Detail: "lpa must be signed"}}
	}

	if lpa.CertificateProvider.SignedAt == nil || lpa.CertificateProvider.SignedAt.IsZero() {
		return []shared.FieldError{{Source: "/type", Detail: "lpa must have a certificate"}}
	}

	for _, a := range lpa.Attorneys {
		if a.Status != shared.AttorneyStatusRemoved && (a.SignedAt == nil || a.SignedAt.IsZero()) {
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

	lpa.Status = shared.LpaStatusStatutoryWaitingPeriod

	return nil
}

func validateStatutoryWaitingPeriod(changes []shared.Change) (StatutoryWaitingPeriod, []shared.FieldError) {
	if len(changes) > 0 {
		return StatutoryWaitingPeriod{}, []shared.FieldError{{Source: "/changes", Detail: "expected empty"}}
	}

	return StatutoryWaitingPeriod{}, nil
}
