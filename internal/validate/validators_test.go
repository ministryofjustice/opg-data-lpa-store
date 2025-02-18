package validate

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestNotEmpty(t *testing.T) {
	assert.Equal(t, "", NotEmpty().Valid("a"))
	assert.Equal(t, "field is required", NotEmpty().Valid(""))
	assert.Equal(t, "", NotEmpty().Valid(time.Now()))
	assert.Equal(t, "field is required", NotEmpty().Valid(time.Time{}))
}

func TestEmpty(t *testing.T) {
	assert.Equal(t, "", Empty().Valid(""))
	assert.Equal(t, "field must not be provided", Empty().Valid("a"))
	assert.Equal(t, "", Empty().Valid(time.Time{}))
	assert.Equal(t, "field must not be provided", Empty().Valid(time.Now()))
}

func TestUUID(t *testing.T) {
	assert.Equal(t, "", UUID().Valid("dc487ebb-b39d-45ed-bb6a-7f950fd355c9"))
	assert.Equal(t, "invalid format", UUID().Valid("dc487ebb-b39d-45ed-bb6a-7f950fd355c"))
	assert.Equal(t, "field is required", UUID().Valid(""))
}

func TestDate(t *testing.T) {
	assert.Equal(t, "", Date().Valid(newDate("2010-01-02")))
	assert.Equal(t, "invalid format", Date().Valid(shared.Date{IsMalformed: true}))
	assert.Equal(t, "field is required", Date().Valid(shared.Date{}))
}

func TestOptionalTime(t *testing.T) {
	now := time.Now()
	assert.Equal(t, "", OptionalTime().Valid(&now))
	assert.Equal(t, "must be a valid datetime", OptionalTime().Valid(&time.Time{}))
	assert.Equal(t, "", OptionalTime().Valid((*time.Time)(nil)))
}

func TestAddressInvalidCountry(t *testing.T) {
	assert.Equal(t, "", Country().Valid("GB"))
	assert.Equal(t, "must be a valid ISO-3166-1 country code", Country().Valid("United Kingdom"))
}

type testIsValid string

func (t testIsValid) IsValid() bool { return string(t) == "ok" }

func TestIsValid(t *testing.T) {
	assert.Equal(t, "", Valid().Valid(testIsValid("ok")))
	assert.Equal(t, "field is required", Valid().Valid(testIsValid("")))
	assert.Equal(t, "invalid value", Valid().Valid(testIsValid("x")))
}

type testUnset bool

func (t testUnset) Unset() bool { return bool(t) }

func TestUnset(t *testing.T) {
	assert.Equal(t, "", Unset().Valid(testUnset(true)))
	assert.Equal(t, "field must not be provided", Unset().Valid(testUnset(false)))
}
