package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
			Key: "/donor/identityCheck/type",
			Old: json.RawMessage("null"),
			New: json.RawMessage(`"one-login"`),
		},
	}

	idCheckComplete, errors := validateDonorConfirmIdentity(changes, &shared.Lpa{})

	assert.Len(t, errors, 0)
	assert.Equal(t, shared.IdentityCheckTypeOneLogin, idCheckComplete.IdentityCheck.Type)
	assert.Equal(t, today.Format(time.RFC3339Nano), idCheckComplete.IdentityCheck.CheckedAt.Format(time.RFC3339Nano))
	assert.Equal(t, donor, idCheckComplete.Actor)
}

func TestConfirmIdentityDonorBadFieldsFails(t *testing.T) {
	changes := []shared.Change{
		// irrelevant field with no prefix
		{
			Key: "/irrelevant",
			Old: json.RawMessage("null"),
			New: json.RawMessage(`"` + time.Now().Format(time.RFC3339Nano) + `"`),
		},
		// irrelevant field with prefix
		{
			Key: "/donor/identityCheck/irrelevant",
			Old: json.RawMessage("null"),
			New: json.RawMessage(`"` + time.Now().Format(time.RFC3339Nano) + `"`),
		},
		// invalid value for field
		{
			Key: "/donor/identityCheck/type",
			Old: json.RawMessage("null"),
			New: json.RawMessage(`"rinky-dink-login-system"`),
		},
	}

	idCheckComplete, errors := validateDonorConfirmIdentity(changes, &shared.Lpa{})

	assert.Len(t, errors, 4)
	assert.Contains(t, errors, shared.FieldError{Source: "/changes", Detail: "missing /donor/identityCheck/checkedAt"})
	assert.Contains(t, errors, shared.FieldError{Source: "/changes/0", Detail: "unexpected change provided"})
	assert.Contains(t, errors, shared.FieldError{Source: "/changes/1", Detail: "unexpected change provided"})
	assert.Contains(t, errors, shared.FieldError{Source: "/changes/2/new", Detail: "invalid value"})
	assert.Equal(t, &shared.IdentityCheck{Type: "rinky-dink-login-system"}, idCheckComplete.IdentityCheck)
	assert.Equal(t, donor, idCheckComplete.Actor)
}

func TestConfirmIdentityDonorANDCertificateProviderFails(t *testing.T) {
	changes := []shared.Change{
		{
			Key: "/certificateProvider/identityCheck/checkedAt",
			Old: json.RawMessage("null"),
			New: json.RawMessage(`"` + time.Now().Format(time.RFC3339Nano) + `"`),
		},
		{
			Key: "/donor/identityCheck/type",
			Old: json.RawMessage("null"),
			New: json.RawMessage(`"one-login"`),
		},
	}

	idCheckComplete, errors := validateDonorConfirmIdentity(changes, &shared.Lpa{})
	expectedIdCheckComplete := &shared.IdentityCheck{
		Type: shared.IdentityCheckTypeOneLogin,
	}

	assert.Len(t, errors, 2)
	assert.Contains(t, errors, shared.FieldError{Source: "/changes", Detail: "missing /donor/identityCheck/checkedAt"})
	assert.Contains(t, errors, shared.FieldError{Source: "/changes/0", Detail: "unexpected change provided"})

	assert.Equal(t, expectedIdCheckComplete, idCheckComplete.IdentityCheck)
	assert.Equal(t, donor, idCheckComplete.Actor)
}

func TestConfirmIdentityDonorMismatchWithExistingLpaFails(t *testing.T) {
	changes := []shared.Change{
		{
			Key: "/donor/identityCheck/checkedAt",
			Old: json.RawMessage("null"),
			New: json.RawMessage(`"` + time.Now().Format(time.RFC3339Nano) + `"`),
		},
		{
			Key: "/donor/identityCheck/type",
			Old: json.RawMessage("null"),
			New: json.RawMessage(`"one-login"`),
		},
	}

	existingLpa := &shared.Lpa{
		LpaInit: shared.LpaInit{
			Donor: shared.Donor{
				IdentityCheck: &shared.IdentityCheck{
					CheckedAt: time.Now().AddDate(-1, 0, 0),
					Type:      "not-one-login",
				},
			},
		},
	}

	idCheckComplete, errors := validateDonorConfirmIdentity(changes, existingLpa)

	assert.ElementsMatch(t, []shared.FieldError{
		{Source: "/changes/0/old", Detail: "does not match existing value"},
		{Source: "/changes/1/old", Detail: "does not match existing value"},
	}, errors)
	assert.Equal(t, existingLpa.Donor.IdentityCheck, idCheckComplete.IdentityCheck)
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
			Key: "/certificateProvider/identityCheck/type",
			Old: json.RawMessage("null"),
			New: json.RawMessage(`"opg-paper-id"`),
		},
	}

	idCheckComplete, errors := validateCertificateProviderConfirmIdentity(changes, &shared.Lpa{})

	assert.Len(t, errors, 0)
	assert.Equal(t, shared.IdentityCheckTypeOpgPaperId, idCheckComplete.IdentityCheck.Type)
	assert.Equal(t, today.Format(time.RFC3339Nano), idCheckComplete.IdentityCheck.CheckedAt.Format(time.RFC3339Nano))
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
