package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

func TestConfirmIdentityDonor(t *testing.T) {
	today := time.Now()

	changes := []shared.Change{
		{
			Key: "/donor/identityCheck/checkedAt",
			Old: json.RawMessage("null"),
			New: json.RawMessage(`"` + today.Format(time.RFC3339Nano) + `"`),
		},
		{
			Key: "/donor/identityCheck/reference",
			Old: json.RawMessage("null"),
			New: json.RawMessage(`"xyz"`),
		},
		{
			Key: "/donor/identityCheck/type",
			Old: json.RawMessage("null"),
			New: json.RawMessage(`"one-login"`),
		},
	}

	idCheckComplete, errors := validateDonorConfirmIdentity(changes, &shared.Lpa{})

	assert.Len(t, errors, 0)
	assert.Equal(t, "xyz", idCheckComplete.Reference)
	assert.Equal(t, shared.IdentityCheckTypeOneLogin, idCheckComplete.Type)
	assert.Equal(t, today.Format(time.RFC3339Nano), idCheckComplete.CheckedAt.Format(time.RFC3339Nano))
	assert.Equal(t, donor, idCheckComplete.Actor)
}

func TestConfirmIdentityCertificateProvider(t *testing.T) {
	today := time.Now()

	changes := []shared.Change{
		{
			Key: "/certificateProvider/identityCheck/checkedAt",
			Old: json.RawMessage("null"),
			New: json.RawMessage(`"` + today.Format(time.RFC3339Nano) + `"`),
		},
		{
			Key: "/certificateProvider/identityCheck/reference",
			Old: json.RawMessage("null"),
			New: json.RawMessage(`"abn"`),
		},
		{
			Key: "/certificateProvider/identityCheck/type",
			Old: json.RawMessage("null"),
			New: json.RawMessage(`"opg-paper-id"`),
		},
	}

	idCheckComplete, errors := validateCertificateProviderConfirmIdentity(changes, &shared.Lpa{})

	assert.Len(t, errors, 0)
	assert.Equal(t, "abn", idCheckComplete.Reference)
	assert.Equal(t, shared.IdentityCheckTypeOpgPaperId, idCheckComplete.Type)
	assert.Equal(t, today.Format(time.RFC3339Nano), idCheckComplete.CheckedAt.Format(time.RFC3339Nano))
	assert.Equal(t, certificateProvider, idCheckComplete.Actor)
}

func TestConfirmIdentityApplyDonor(t *testing.T) {
	check := IdCheckComplete{
		Actor:         donor,
		IdentityCheck: &shared.IdentityCheck{},
	}

	lpa := shared.Lpa{}

	check.Apply(&lpa)

	assert.Equal(t, check.IdentityCheck, lpa.Donor.IdentityCheck)
}

func TestConfirmIdentityApplyCertificateProvider(t *testing.T) {
	check := IdCheckComplete{
		Actor:         certificateProvider,
		IdentityCheck: &shared.IdentityCheck{},
	}

	lpa := shared.Lpa{}

	check.Apply(&lpa)

	assert.Equal(t, check.IdentityCheck, lpa.CertificateProvider.IdentityCheck)
}
