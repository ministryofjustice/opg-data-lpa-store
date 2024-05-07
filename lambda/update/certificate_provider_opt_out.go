package main

import "github.com/ministryofjustice/opg-data-lpa-store/internal/shared"

type CertificateProviderOptOut struct{}

func (c CertificateProviderOptOut) Apply(lpa *shared.Lpa) []shared.FieldError {
	emptyCertificateProvider := shared.CertificateProvider{}
	if lpa.CertificateProvider == emptyCertificateProvider {
		return []shared.FieldError{{Source: "/type", Detail: "certificate provider not present on LPA"}}
	}

	if lpa.CertificateProvider.SignedAt != nil && !lpa.CertificateProvider.SignedAt.IsZero() {
		return []shared.FieldError{{Source: "/type", Detail: "certificate provider cannot opt out after providing certificate"}}
	}

	lpa.CertificateProviderNotRelatedConfirmedAt = nil
	lpa.CertificateProvider = shared.CertificateProvider{}

	return nil
}

func validateCertificateProviderOptOut(changes []shared.Change) (CertificateProviderOptOut, []shared.FieldError) {
	if len(changes) > 0 {
		return CertificateProviderOptOut{}, []shared.FieldError{{Source: "/changes", Detail: "expected empty"}}
	}

	return CertificateProviderOptOut{}, nil
}
