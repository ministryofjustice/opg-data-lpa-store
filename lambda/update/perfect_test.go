package main

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestPerfectApply(t *testing.T) {
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

	errors := Perfect{}.Apply(lpa)
	assert.Nil(t, errors)
	assert.Equal(t, shared.LpaStatusPerfect, lpa.Status)
}

func TestPerfectApplyWhenUnsigned(t *testing.T) {
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
			errors := Perfect{}.Apply(tc.lpa)
			assert.Equal(t, tc.errors, errors)
		})
	}
}

func TestRegisterApplyWhenNotInProgress(t *testing.T) {
	for _, status := range []shared.LpaStatus{shared.LpaStatusPerfect, shared.LpaStatusRegistered} {
		t.Run(string(status), func(t *testing.T) {
			errors := Perfect{}.Apply(&shared.Lpa{Status: status})
			assert.Equal(t, []shared.FieldError{{Source: "/type", Detail: "status must be in-progress to make perfect"}}, errors)
		})
	}
}

func TestValidatePerfect(t *testing.T) {
	_, errors := validatePerfect(nil)
	assert.Nil(t, errors)
}

func TestValidatePerfectWhenChanges(t *testing.T) {
	_, errors := validatePerfect([]shared.Change{{}})
	assert.Equal(t, []shared.FieldError{{Source: "/changes", Detail: "expected empty"}}, errors)
}
