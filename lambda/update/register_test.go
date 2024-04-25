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
		Status: shared.LpaStatusPerfect,
	}

	errors := Register{}.Apply(lpa)
	assert.Nil(t, errors)
	assert.WithinDuration(t, now, *lpa.RegistrationDate, time.Millisecond)
	assert.Equal(t, shared.LpaStatusRegistered, lpa.Status)
}

func TestRegisterApplyWhenNotPerfect(t *testing.T) {
	for _, status := range []shared.LpaStatus{shared.LpaStatusProcessing, shared.LpaStatusRegistered} {
		t.Run(string(status), func(t *testing.T) {
			errors := Register{}.Apply(&shared.Lpa{Status: status})
			assert.Equal(t, []shared.FieldError{{Source: "/type", Detail: "status must be perfect to register"}}, errors)
		})
	}
}

func TestValidateRegister(t *testing.T) {
	_, errors := validateRegister(nil)
	assert.Nil(t, errors)
}

func TestValidateRegisterWhenChanges(t *testing.T) {
	_, errors := validateRegister([]shared.Change{{}})
	assert.Equal(t, []shared.FieldError{{Source: "/changes", Detail: "expected empty"}}, errors)
}
