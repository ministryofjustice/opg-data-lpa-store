package main

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestStatutoryWaitingPeriodApply(t *testing.T) {
	now := time.Now()

	lpa := &shared.Lpa{
		Status: shared.LpaStatusInProgress,
		LpaInit: shared.LpaInit{
			SignedAt: now,
			CertificateProvider: shared.CertificateProvider{
				SignedAt: &now,
			},
			Attorneys: []shared.Attorney{{
				SignedAt: &now,
			}},
			TrustCorporations: []shared.TrustCorporation{{
				Signatories: []shared.Signatory{{
					SignedAt: now,
				}},
			}},
		},
	}

	errors := StatutoryWaitingPeriod{}.Apply(lpa)
	assert.Nil(t, errors)
	assert.Equal(t, shared.LpaStatusStatutoryWaitingPeriod, lpa.Status)
}

func TestStatutoryWaitingPeriodApplyWhenUnsigned(t *testing.T) {
	now := time.Now()

	testcases := map[string]struct {
		lpa    *shared.Lpa
		errors []shared.FieldError
	}{
		"lpa": {
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					CertificateProvider: shared.CertificateProvider{SignedAt: &now},
					Attorneys:           []shared.Attorney{{SignedAt: &now}},
					TrustCorporations: []shared.TrustCorporation{{
						Signatories: []shared.Signatory{{SignedAt: now}},
					}},
				},
			},
			errors: []shared.FieldError{{Source: "/type", Detail: "lpa must be signed"}},
		},
		"certificate provider": {
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					SignedAt:            now,
					CertificateProvider: shared.CertificateProvider{},
					Attorneys:           []shared.Attorney{{SignedAt: &now}},
					TrustCorporations: []shared.TrustCorporation{{
						Signatories: []shared.Signatory{{SignedAt: now}},
					}},
				},
			},
			errors: []shared.FieldError{{Source: "/type", Detail: "lpa must have a certificate"}},
		},
		"attorney": {
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					SignedAt:            now,
					CertificateProvider: shared.CertificateProvider{SignedAt: &now},
					Attorneys:           []shared.Attorney{{SignedAt: &now}, {}},
					TrustCorporations: []shared.TrustCorporation{{
						Signatories: []shared.Signatory{{SignedAt: now}},
					}},
				},
			},
			errors: []shared.FieldError{{Source: "/type", Detail: "lpa must be signed by attorneys"}},
		},
		"trust corporation": {
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					SignedAt:            now,
					CertificateProvider: shared.CertificateProvider{SignedAt: &now},
					Attorneys:           []shared.Attorney{{SignedAt: &now}},
					TrustCorporations: []shared.TrustCorporation{{
						Signatories: []shared.Signatory{{SignedAt: now}, {}},
					}},
				},
			},
			errors: []shared.FieldError{{Source: "/type", Detail: "lpa must be signed by trust corporations"}},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			errors := StatutoryWaitingPeriod{}.Apply(tc.lpa)
			assert.Equal(t, tc.errors, errors)
		})
	}
}

func TestRegisterApplyWhenNotInProgress(t *testing.T) {
	for _, status := range []shared.LpaStatus{shared.LpaStatusStatutoryWaitingPeriod, shared.LpaStatusRegistered} {
		t.Run(string(status), func(t *testing.T) {
			errors := StatutoryWaitingPeriod{}.Apply(&shared.Lpa{Status: status})
			assert.Equal(t, []shared.FieldError{{Source: "/type", Detail: "status must be in-progress to make statutory waiting period"}}, errors)
		})
	}
}

func TestValidateStatutoryWaitingPeriod(t *testing.T) {
	_, errors := validateStatutoryWaitingPeriod(nil)
	assert.Nil(t, errors)
}

func TestValidateStatutoryWaitingPeriodWhenChanges(t *testing.T) {
	_, errors := validateStatutoryWaitingPeriod([]shared.Change{{}})
	assert.Equal(t, []shared.FieldError{{Source: "/changes", Detail: "expected empty"}}, errors)
}
