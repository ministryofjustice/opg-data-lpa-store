package parse

import (
	"encoding/json"
	"testing"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

var jsonNull = json.RawMessage("null")

func TestField(t *testing.T) {
	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	var v string
	Changes(changes).Field("/thing", &v)

	assert.Equal(t, "val", v)
}

func TestFieldWhenMissing(t *testing.T) {
	changes := []shared.Change{}

	var v string
	errors := Changes(changes).Field("/thing", &v).Errors()

	assert.Equal(t, []shared.FieldError{{Source: "/changes", Detail: "missing /thing"}}, errors)
}

func TestFieldWhenWrongType(t *testing.T) {
	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	var v int
	errors := Changes(changes).Field("/thing", &v).Errors()

	assert.Equal(t, []shared.FieldError{{Source: "/changes/0/new", Detail: "wrong type"}}, errors)
}

func TestFieldOptional(t *testing.T) {
	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	var v string
	Changes(changes).Field("/thing", &v, Optional())

	assert.Equal(t, "val", v)
}

func TestFieldOptionalWhenMissing(t *testing.T) {
	changes := []shared.Change{}

	var v string
	errors := Changes(changes).Field("/thing", &v, Optional()).Errors()

	assert.Empty(t, errors)
}

func TestFieldValidate(t *testing.T) {
	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	var v string
	Changes(changes).Field("/thing", &v, Validate(func() []shared.FieldError {
		return nil
	}))

	assert.Equal(t, "val", v)
}

func TestFieldValidateWhenInvalid(t *testing.T) {
	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"what"`), Old: jsonNull},
	}

	var v string
	errors := Changes(changes).Field("/thing", &v, Validate(func() []shared.FieldError {
		return []shared.FieldError{{Source: "/rewritten", Detail: "invalid"}}
	})).Errors()

	assert.Equal(t, []shared.FieldError{{Source: "/changes/0/new", Detail: "invalid"}}, errors)
}

func TestConsumed(t *testing.T) {
	changes := []shared.Change{}
	errors := Changes(changes).Consumed()

	assert.Empty(t, errors)
}

func TestConsumedWhenNot(t *testing.T) {
	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	errors := Changes(changes).Consumed()

	assert.Equal(t, []shared.FieldError{{Source: "/changes/0", Detail: "unexpected change provided"}}, errors)
}

func TestConsumedWhenConsumed(t *testing.T) {
	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	var v string
	errors := Changes(changes).Field("/thing", &v).Consumed()

	assert.Empty(t, errors)
}

func TestEach(t *testing.T) {
	changes := []shared.Change{
		{Key: "/0/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
		{Key: "/1/other", New: json.RawMessage(`"other"`), Old: jsonNull},
	}

	var v, w string
	errors := Changes(changes).Each(func(i int, p *Parser) []shared.FieldError {
		if i == 0 {
			p.Field("/thing", &v)
		} else if i == 1 {
			p.Field("/other", &w)
		}
		return p.Consumed()
	}).Errors()

	assert.Equal(t, "val", v)
	assert.Equal(t, "other", w)
	assert.Empty(t, errors)
}

func TestEachWhenNonIndexedKey(t *testing.T) {
	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
		{Key: "/-/other", New: json.RawMessage(`"other"`), Old: jsonNull},
	}

	errors := Changes(changes).Each(func(i int, p *Parser) []shared.FieldError {
		var v any
		p.Field("/thing", v)
		return p.Errors()
	}).Errors()

	assert.Equal(t, []shared.FieldError{
		{Source: "/changes/0/key", Detail: "require index"},
		{Source: "/changes/1/key", Detail: "require index"},
	}, errors)
}

func TestEachWhenRequired(t *testing.T) {
	changes := []shared.Change{}

	errors := Changes(changes).Each(func(i int, p *Parser) []shared.FieldError {
		var v any
		p.Field("/thing", v)
		return p.Errors()
	}, 0).Errors()

	assert.Equal(t, []shared.FieldError{
		{Source: "/changes", Detail: "missing /0/thing"},
	}, errors)
}

func TestEachWhenOutOfRange(t *testing.T) {
	changes := []shared.Change{
		{Key: "/0/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
		{Key: "/1/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
		{Key: "/2/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	errors := Changes(changes).Each(func(i int, p *Parser) []shared.FieldError {
		if i > 0 {
			return p.OutOfRange()
		}

		return p.Errors()
	}).Errors()

	assert.ElementsMatch(t, []shared.FieldError{
		{Source: "/changes/1/key", Detail: "index out of range"},
		{Source: "/changes/2/key", Detail: "index out of range"},
	}, errors)
}

func TestEachWhenNotConsumed(t *testing.T) {
	changes := []shared.Change{
		{Key: "/0/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
		{Key: "/1/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
		{Key: "/2/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	errors := Changes(changes).Each(func(i int, p *Parser) []shared.FieldError {
		if i == 0 {
			var v string
			p.Field("/thing", &v)
		}

		return p.Consumed()
	}).Errors()

	assert.ElementsMatch(t, []shared.FieldError{
		{Source: "/changes/1", Detail: "unexpected change provided"},
		{Source: "/changes/2", Detail: "unexpected change provided"},
	}, errors)
}

func TestPrefix(t *testing.T) {
	changes := []shared.Change{
		{Key: "/a/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	var v string
	errors := Changes(changes).Prefix("/a", func(p *Parser) []shared.FieldError {
		return p.
			Field("/thing", &v).
			Consumed()
	}).Consumed()

	assert.Equal(t, "val", v)
	assert.Empty(t, errors)
}

func TestPrefixWhenNotConsumed(t *testing.T) {
	changes := []shared.Change{
		{Key: "/a/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
		{Key: "/a/what", New: json.RawMessage(`"val"`), Old: jsonNull},
		{Key: "/root", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	var v string
	errors := Changes(changes).Prefix("/a", func(p *Parser) []shared.FieldError {
		return p.
			Field("/thing", &v).
			Consumed()
	}).Consumed()

	assert.Equal(t, "val", v)
	assert.ElementsMatch(t, []shared.FieldError{
		{Source: "/changes/1", Detail: "unexpected change provided"},
		{Source: "/changes/2", Detail: "unexpected change provided"},
	}, errors)
}

func TestPrefixWhenMissing(t *testing.T) {
	changes := []shared.Change{}

	errors := Changes(changes).Prefix("/a", func(p *Parser) []shared.FieldError { return nil }).Consumed()

	assert.ElementsMatch(t, []shared.FieldError{
		{Source: "/changes", Detail: "missing /a/..."},
	}, errors)
}

func TestOptionalPrefix(t *testing.T) {
	changes := []shared.Change{
		{Key: "/a/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	var v string
	errors := Changes(changes).OptionalPrefix("/a", func(p *Parser) []shared.FieldError {
		return p.
			Field("/thing", &v).
			Consumed()
	}).Consumed()

	assert.Equal(t, "val", v)
	assert.Empty(t, errors)
}

func TestOptionalPrefixWhenNotConsumed(t *testing.T) {
	changes := []shared.Change{
		{Key: "/a/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
		{Key: "/a/what", New: json.RawMessage(`"val"`), Old: jsonNull},
		{Key: "/root", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	var v string
	errors := Changes(changes).OptionalPrefix("/a", func(p *Parser) []shared.FieldError {
		return p.Field("/thing", &v).Consumed()
	}).Consumed()

	assert.Equal(t, "val", v)
	assert.ElementsMatch(t, []shared.FieldError{
		{Source: "/changes/1", Detail: "unexpected change provided"},
		{Source: "/changes/2", Detail: "unexpected change provided"},
	}, errors)
}

func TestOptionalPrefixWhenMissing(t *testing.T) {
	changes := []shared.Change{}
	errors := Changes(changes).OptionalPrefix("/a", func(p *Parser) []shared.FieldError { return nil }).Consumed()

	assert.Empty(t, errors)
}
