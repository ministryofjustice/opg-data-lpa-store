package main

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestCertificateProviderOptOutApply(t *testing.T) {
	lpa := &shared.Lpa{LpaInit: shared.LpaInit{CertificateProvider: shared.CertificateProvider{Email: "a@example"}}}
	c := CertificateProviderOptOut{}

	errors := c.Apply(lpa)

	assert.Empty(t, errors)
	assert.Equal(t, shared.CertificateProvider{}, lpa.CertificateProvider)
	assert.Nil(t, lpa.CertificateProviderNotRelatedConfirmedAt)
}

func TestCertificateProviderOptOutApplyWhenNoCertificateProvider(t *testing.T) {
	lpa := &shared.Lpa{LpaInit: shared.LpaInit{
		CertificateProvider: shared.CertificateProvider{}},
	}

	errors := CertificateProviderOptOut{}.Apply(lpa)

	assert.Equal(t, errors, []shared.FieldError{{Source: "/type", Detail: "certificate provider not present on LPA"}})
}

func TestCertificateProviderOptOutApplyWhenCertificateProvided(t *testing.T) {
	now := time.Now()
	certificateProvider := shared.CertificateProvider{Email: "a@example", SignedAt: &now}
	lpa := &shared.Lpa{LpaInit: shared.LpaInit{
		CertificateProvider: certificateProvider},
	}

	errors := CertificateProviderOptOut{}.Apply(lpa)

	assert.Equal(t, errors, []shared.FieldError{{Source: "/type", Detail: "certificate provider cannot opt out after providing certificate"}})
	assert.Equal(t, certificateProvider, lpa.CertificateProvider)
}
