package main

import (
	"testing"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestValidateDonorWithdrawLPA(t *testing.T) {
	_, errors := validateDonorWithdrawLPA([]shared.Change{})
	assert.Nil(t, errors)
}

func TestValidateDonorWithdrawLPAWithChanges(t *testing.T) {
	_, errors := validateDonorWithdrawLPA([]shared.Change{{}})
	assert.Equal(t, []shared.FieldError{{Source: "/changes", Detail: "expected empty"}}, errors)
}

func TestDonorWithdrawLPA(t *testing.T) {
	lpa := &shared.Lpa{
		Status: shared.LpaStatusInProgress,
	}

	errors := DonorWithdrawLpa{}.Apply(lpa)
	assert.Nil(t, errors)
	assert.Equal(t, shared.LpaStatusWithdrawn, lpa.Status)
}

func TestDonorWithdrawLPAInvalidStatuses(t *testing.T) {
	testcases := map[shared.LpaStatus]string{
		shared.LpaStatusWithdrawn:      "lpa has already been withdrawn",
		shared.LpaStatusRegistered:     "cannot withdraw a registered lpa",
		shared.LpaStatusCannotRegister: "cannot withdraw an unregisterable lpa",
	}

	for status, expectedError := range testcases {
		t.Run(string(status), func(t *testing.T) {
			errors := DonorWithdrawLpa{}.Apply(&shared.Lpa{Status: status})

			assert.Equal(t, []shared.FieldError{{Source: "/type", Detail: expectedError}}, errors)
		})
	}
}
