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
		{Key: "/invalid", Old: json.RawMessage(`not-json`), New: json.RawMessage(`null`)},
	}

	errors := redundantChangeErrors(changes)

	assert.Equal(t, []shared.FieldError{{
		Source: "/changes/0",
		Detail: "redundant change for /duplicate",
	}}, errors)
}

func TestRedundantChangeErrorsEmpty(t *testing.T) {
	assert.Nil(t, redundantChangeErrors(nil))
	assert.Nil(t, redundantChangeErrors([]shared.Change{}))
}

func TestIsRedundantChangeNilAndEmpty(t *testing.T) {
	redundant, err := isRedundantChange(json.RawMessage(`null`), json.RawMessage(`""`))
	assert.NoError(t, err)
	assert.True(t, redundant)

	redundant, err = isRedundantChange(json.RawMessage(`""`), json.RawMessage(`null`))
	assert.NoError(t, err)
	assert.True(t, redundant)
}

func TestIsRedundantChangeNormalisesValues(t *testing.T) {
	t.Run("numbers", func(t *testing.T) {
		redundant, err := isRedundantChange(json.RawMessage(`1`), json.RawMessage(`1.0`))
		assert.NoError(t, err)
		assert.True(t, redundant)
	})

	t.Run("objects", func(t *testing.T) {
		redundant, err := isRedundantChange(json.RawMessage(`{"a":1,"b":[true,"x"]}`), json.RawMessage(`{"b":[true,"x"],"a":1}`))
		assert.NoError(t, err)
		assert.True(t, redundant)
	})
}

func TestIsRedundantChangeInvalidJSON(t *testing.T) {
	redundant, err := isRedundantChange(json.RawMessage(`not-json`), json.RawMessage(`null`))
	assert.Error(t, err)
	assert.False(t, redundant)
}
