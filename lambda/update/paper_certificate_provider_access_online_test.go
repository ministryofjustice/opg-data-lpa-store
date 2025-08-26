package main

import (
	"encoding/json"
	"testing"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestValidateUpdatePaperCertificateProviderAccessOnline(t *testing.T) {
	testcases := map[string]struct {
		update shared.Update
		errors []shared.FieldError
		email  string
	}{
		"valid": {
			update: shared.Update{
				Type: "PAPER_CERTIFICATE_PROVIDER_ACCESS_ONLINE",
				Changes: []shared.Change{
					{Key: "/certificateProvider/email", Old: jsonNull, New: json.RawMessage(`"a@example.com"`)},
				},
			},
		},
		"missing email": {
			update: shared.Update{
				Type: "PAPER_CERTIFICATE_PROVIDER_ACCESS_ONLINE",
				Changes: []shared.Change{
					{Key: "/certificateProvider/email", Old: jsonNull, New: jsonNull},
				},
			},
			errors: []shared.FieldError{
				{Source: "/changes/0/new", Detail: "field is required"},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			_, errors := validateUpdate(tc.update, &shared.Lpa{LpaInit: shared.LpaInit{CertificateProvider: shared.CertificateProvider{Email: tc.email}}})
			assert.ElementsMatch(t, tc.errors, errors)
		})
	}
}

func TestPaperCertificateProviderAccessOnlineApply(t *testing.T) {
	lpa := shared.Lpa{LpaInit: shared.LpaInit{CertificateProvider: shared.CertificateProvider{Channel: shared.ChannelPaper}}}
	errors := PaperCertificateProviderAccessOnline{Email: "a@example.com"}.Apply(&lpa)

	assert.Len(t, errors, 0)
	assert.Equal(t, "a@example.com", lpa.CertificateProvider.Email)
}

func TestPaperCertificateProviderAccessOnlineApplyWhenChannelNotPaper(t *testing.T) {
	lpa := shared.Lpa{LpaInit: shared.LpaInit{CertificateProvider: shared.CertificateProvider{Channel: shared.ChannelOnline}}}
	errors := PaperCertificateProviderAccessOnline{Email: "a@example.com"}.Apply(&lpa)

	assert.Len(t, errors, 1)
	assert.Equal(t, shared.FieldError{Source: "/channel", Detail: "lpa channel is not paper"}, errors[0])
	assert.Equal(t, "", lpa.CertificateProvider.Email)
}
