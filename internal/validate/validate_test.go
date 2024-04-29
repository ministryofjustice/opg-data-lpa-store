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

func TestRequired(t *testing.T) {
	assert.Nil(t, Required("a", "a"))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "field is required"}}, Required("a", ""))
}

func TestEmpty(t *testing.T) {
	assert.Nil(t, Empty("a", ""))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "field must not be provided"}}, Empty("a", "a"))
}

func TestUUID(t *testing.T) {
	assert.Nil(t, UUID("a", "dc487ebb-b39d-45ed-bb6a-7f950fd355c9"))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "invalid format"}}, UUID("a", "dc487ebb-b39d-45ed-bb6a-7f950fd355c"))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "field is required"}}, UUID("a", ""))
}

func TestDate(t *testing.T) {
	assert.Nil(t, Date("a", newDate("2010-01-02")))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "invalid format"}}, Date("a", shared.Date{IsMalformed: true}))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "field is required"}}, Date("a", shared.Date{}))
}

func TestTime(t *testing.T) {
	assert.Nil(t, Time("a", time.Now()))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "field is required"}}, Time("a", time.Time{}))
}

func TestOptionalTime(t *testing.T) {
	now := time.Now()
	assert.Nil(t, OptionalTime("a", &now))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "must be a valid datetime"}}, OptionalTime("a", &time.Time{}))
	assert.Nil(t, OptionalTime("a", nil))
}

func TestAddressEmpty(t *testing.T) {
	address := shared.Address{}
	errors := Address("/test", address)

	assert.Contains(t, errors, shared.FieldError{Source: "/test/line1", Detail: "field is required"})
	assert.Contains(t, errors, shared.FieldError{Source: "/test/country", Detail: "field is required"})
}

func TestAddressValid(t *testing.T) {
	errors := Address("/test", validAddress)

	assert.Empty(t, errors)
}

func TestAddressInvalidCountry(t *testing.T) {
	invalidAddress := shared.Address{
		Line1:   "123 Main St",
		Country: "United Kingdom",
	}
	errors := Address("/test", invalidAddress)

	assert.Contains(t, errors, shared.FieldError{Source: "/test/country", Detail: "must be a valid ISO-3166-1 country code"})
}

type testIsValid string

func (t testIsValid) IsValid() bool { return string(t) == "ok" }

func TestIsValid(t *testing.T) {
	assert.Nil(t, IsValid("a", testIsValid("ok")))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "field is required"}}, IsValid("a", testIsValid("")))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "invalid value"}}, IsValid("a", testIsValid("x")))
}

type testUnset bool

func (t testUnset) Unset() bool { return bool(t) }

func TestUnset(t *testing.T) {
	assert.Nil(t, Unset("a", testUnset(true)))
	assert.Equal(t, []shared.FieldError{{Source: "a", Detail: "field must not be provided"}}, Unset("a", testUnset(false)))
}
