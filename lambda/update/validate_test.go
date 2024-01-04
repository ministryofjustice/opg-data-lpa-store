package main

import (
	"testing"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestValidateUpdate(t *testing.T) {
	applyable, errors := validateUpdate(shared.Update{Type: "what"})
	assert.Nil(t, applyable)
	assert.Equal(t, []shared.FieldError{{Source: "/type", Detail: "invalid value"}}, errors)
}

func TestErrorMustBeString(t *testing.T) {
	assert.Equal(t, []shared.FieldError{{Source: "x", Detail: "must be a string"}}, errorMustBeString([]shared.FieldError{}, "x"))
}
