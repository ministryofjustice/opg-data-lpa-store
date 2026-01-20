package main

import (
	"encoding/json"
	"testing"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestRedundantChangeErrors(t *testing.T) {
	changes := []shared.Change{
		{Key: "/duplicate", Old: json.RawMessage(`"foo"`), New: json.RawMessage(`"foo"`)},
		{Key: "/different", Old: json.RawMessage(`"foo"`), New: json.RawMessage(`"bar"`)},
	}

	errors, err := redundantChangeErrors(changes)
	assert.NoError(t, err)

	assert.Equal(t, []shared.FieldError{{
		Source: "/changes/0",
		Detail: "redundant change for /duplicate",
	}}, errors)
}

func TestRedundantChangeErrorsEmpty(t *testing.T) {
	errors, err := redundantChangeErrors(nil)
	assert.NoError(t, err)
	assert.Nil(t, errors)

	errors, err = redundantChangeErrors([]shared.Change{})
	assert.NoError(t, err)
	assert.Nil(t, errors)
}

func TestRedundantChangeErrorsInvalidJSON(t *testing.T) {
	changes := []shared.Change{{Key: "/invalid", Old: json.RawMessage(`not-json`), New: json.RawMessage(`null`)}}

	errors, err := redundantChangeErrors(changes)
	assert.Error(t, err)
	assert.Nil(t, errors)
}

func TestIsRedundantChangeNilAndEmpty(t *testing.T) {
	redundant, err := isRedundantChange(json.RawMessage(`null`), json.RawMessage(`""`))
	assert.NoError(t, err)
	assert.True(t, redundant)

	redundant, err = isRedundantChange(json.RawMessage(`""`), json.RawMessage(`null`))
	assert.NoError(t, err)
	assert.True(t, redundant)
}
func TestIsRedundantChangeInvalidJSON(t *testing.T) {
	redundant, err := isRedundantChange(json.RawMessage(`not-json`), json.RawMessage(`null`))
	assert.Error(t, err)
	assert.False(t, redundant)
}
