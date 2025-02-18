package validate

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

var validAddress = shared.Address{
	Line1:   "123 Main St",
	Country: "GB",
}

func newDate(date string) shared.Date {
	d := shared.Date{}
	_ = d.UnmarshalText([]byte(date))
	return d
}

func TestAll(t *testing.T) {
	errA := shared.FieldError{Source: "a", Detail: "a"}
	errB := shared.FieldError{Source: "b", Detail: "b"}
	errC := shared.FieldError{Source: "c", Detail: "c"}

	assert.Nil(t, All())
	assert.Nil(t, All([]shared.FieldError{}, []shared.FieldError{}))
	assert.Equal(t, []shared.FieldError{errA, errB, errC}, All([]shared.FieldError{errA, errB}, []shared.FieldError{errC}))
	assert.Equal(t, []shared.FieldError{errA, errB, errC}, All([]shared.FieldError{errA}, []shared.FieldError{errB, errC}))
	assert.Equal(t, []shared.FieldError{errA, errB, errC}, All([]shared.FieldError{errA}, []shared.FieldError{errB}, []shared.FieldError{errC}))
}

func TestIf(t *testing.T) {
	errs := []shared.FieldError{{Source: "a", Detail: "a"}}

	assert.Equal(t, errs, If(true, errs))
	assert.Nil(t, If(false, errs))
}

func TestIfElse(t *testing.T) {
	errsA := []shared.FieldError{{Source: "a", Detail: "a"}}
	errsB := []shared.FieldError{{Source: "b", Detail: "b"}}

	assert.Equal(t, errsA, IfElse(true, errsA, errsB))
	assert.Equal(t, errsB, IfElse(false, errsA, errsB))
}

func TestNotEmpty(t *testing.T) {
	assert.Equal(t, "", NotEmpty().Valid("a"))
	assert.Equal(t, "field is required", NotEmpty().Valid(""))
}

func TestEmpty(t *testing.T) {
	assert.Equal(t, "", Empty().Valid(""))
	assert.Equal(t, "field must not be provided", Empty().Valid("a"))
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

func TestTime(t *testing.T) {
	assert.Equal(t, "", NotEmpty().Valid(time.Now()))
	assert.Equal(t, "field is required", NotEmpty().Valid(time.Time{}))
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
	assert.Nil(t, WithSource("a", testUnset(true), Unset()))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "field must not be provided"}}, WithSource("a", testUnset(false), Unset()))
}
