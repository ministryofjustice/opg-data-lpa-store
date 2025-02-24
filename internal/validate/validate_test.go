package validate

import (
	"testing"

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
