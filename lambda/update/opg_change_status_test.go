package main

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpgChangeStatusToCannotRegisterApply(t *testing.T) {
	lpa := &shared.Lpa{
		Status: shared.LpaStatusInProgress,
	}
	c := OpgChangeStatus{
		Status: shared.LpaStatusCannotRegister,
	}

	errors := c.Apply(lpa)
	assert.Empty(t, errors)
	assert.Equal(t, c.Status, lpa.Status)
}

func TestOpgChangeStatusToCancelledApply(t *testing.T) {
	lpa := &shared.Lpa{
		Status: shared.LpaStatusRegistered,
	}
	c := OpgChangeStatus{
		Status: shared.LpaStatusCancelled,
	}

	errors := c.Apply(lpa)
	assert.Empty(t, errors)
	assert.Equal(t, c.Status, lpa.Status)
}

func TestOpgChangeStatusToDoNotRegisterApply(t *testing.T) {
	lpa := &shared.Lpa{
		Status: shared.LpaStatusStatutoryWaitingPeriod,
	}
	c := OpgChangeStatus{
		Status: shared.LpaStatusDoNotRegister,
	}

	errors := c.Apply(lpa)
	assert.Empty(t, errors)
	assert.Equal(t, c.Status, lpa.Status)
}

func TestOpgChangeStatusToExpiredApply(t *testing.T) {
	lpa := &shared.Lpa{
		Status: shared.LpaStatusStatutoryWaitingPeriod,
	}
	c := OpgChangeStatus{
		Status: shared.LpaStatusExpired,
	}

	errors := c.Apply(lpa)
	assert.Empty(t, errors)
	assert.Equal(t, c.Status, lpa.Status)
}

func TestOpgChangeStatusInvalidNewStatus(t *testing.T) {
	lpa := &shared.Lpa{
		Status: shared.LpaStatusInProgress,
	}
	c := OpgChangeStatus{
		Status: shared.LpaStatusWithdrawn,
	}

	errors := c.Apply(lpa)
	assert.Equal(t, errors, []shared.FieldError{{Source: "/status", Detail: "Status to be updated should be cannot register, cancelled, do not register or expired"}})
}

func TestOpgChangeStatusToCannotRegisterIncorrectExistingStatus(t *testing.T) {
	lpa := &shared.Lpa{
		Status: shared.LpaStatusRegistered,
	}
	c := OpgChangeStatus{
		Status: shared.LpaStatusCannotRegister,
	}

	errors := c.Apply(lpa)
	assert.Equal(t, errors, []shared.FieldError{{Source: "/status", Detail: "Lpa status cannot be registered while changing to cannot register"}})
}

func TestOpgChangeStatusToCancelledIncorrectExistingStatus(t *testing.T) {
	lpa := &shared.Lpa{
		Status: shared.LpaStatusInProgress,
	}
	c := OpgChangeStatus{
		Status: shared.LpaStatusCancelled,
	}

	errors := c.Apply(lpa)
	assert.Equal(t, errors, []shared.FieldError{{Source: "/status", Detail: "Lpa status has to be registered while changing to cancelled"}})
}

func TestOpgChangeStatusToDoNotRegisterIncorrectExistingStatus(t *testing.T) {
	lpa := &shared.Lpa{
		Status: shared.LpaStatusInProgress,
	}
	c := OpgChangeStatus{
		Status: shared.LpaStatusDoNotRegister,
	}

	errors := c.Apply(lpa)
	assert.Equal(t, errors, []shared.FieldError{{Source: "/status", Detail: "Lpa status has to be statutory waiting period while changing to do not register"}})
}

func TestOpgChangeStatusToExpiredIncorrectExistingStatus(t *testing.T) {
	lpa := &shared.Lpa{
		Status: shared.LpaStatusRegistered,
	}
	c := OpgChangeStatus{
		Status: shared.LpaStatusExpired,
	}

	errors := c.Apply(lpa)
	assert.Equal(t, errors, []shared.FieldError{{Source: "/status", Detail: "Lpa status has to be in progress, statutory waiting period or do not register while changing to expired"}})
}

func TestValidateUpdateOPGChangeStatus(t *testing.T) {
	testcases := map[string]struct {
		update shared.Update
		lpa    *shared.Lpa
		errors []shared.FieldError
	}{
		"valid - with previous values": {
			update: shared.Update{
				Type: "OPG_STATUS_CHANGE",
				Changes: []shared.Change{
					{
						Key: "/status",
						New: json.RawMessage(`"cannot-register"`),
						Old: json.RawMessage(`"in-progress"`),
					},
				},
			},
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
			},
		},
		"invalid status": {
			update: shared.Update{
				Type: "OPG_STATUS_CHANGE",
				Changes: []shared.Change{
					{
						Key: "/status",
						New: json.RawMessage(`"never-register"`),
						Old: json.RawMessage(`"in-progress"`),
					},
				},
			},
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
			},
			errors: []shared.FieldError{
				{Source: "/changes/0/new", Detail: "invalid value"},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			_, errors := validateUpdate(tc.update, tc.lpa)
			assert.ElementsMatch(t, tc.errors, errors)
		})
	}
}
