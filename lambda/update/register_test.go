package main

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestRegisterApply(t *testing.T) {
	now := time.Now()

	lpa := &shared.Lpa{
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

	errors := Register{}.Apply(lpa)
	assert.Nil(t, errors)
	assert.WithinDuration(t, now, *lpa.RegistrationDate, time.Millisecond)
	assert.Equal(t, shared.LpaStatusRegistered, lpa.Status)
}

func TestRegisterApplyWhenUnsigned(t *testing.T) {
	now := time.Now()

	testcases := map[string]struct {
		lpa    *shared.Lpa
		errors []shared.FieldError
	}{
		"lpa": {
			lpa: &shared.Lpa{
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
			errors := Register{}.Apply(tc.lpa)
			assert.Equal(t, tc.errors, errors)
		})
	}
}

func TestRegisterApplyWhenAlreadyRegistered(t *testing.T) {
	now := time.Now()

	lpa := &shared.Lpa{
		RegistrationDate: &now,
	}

	errors := Register{}.Apply(lpa)
	assert.Equal(t, []shared.FieldError{{Source: "/type", Detail: "lpa already registered"}}, errors)
}

func TestValidateRegister(t *testing.T) {
	_, errors := validateRegister(nil)
	assert.Nil(t, errors)
}

func TestValidateRegisterWhenChanges(t *testing.T) {
	_, errors := validateRegister([]shared.Change{{}})
	assert.Equal(t, []shared.FieldError{{Source: "/changes", Detail: "expected empty"}}, errors)
}
