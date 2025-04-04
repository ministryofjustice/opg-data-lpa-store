package parse

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
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

	assert.Equal(t, []shared.FieldError{{Source: "/positionChanges", Detail: "missing /thing"}}, errors)
}

func TestFieldWhenWrongType(t *testing.T) {
	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	var v int
	errors := Changes(changes).Field("/thing", &v).Errors()

	assert.Equal(t, []shared.FieldError{{Source: "/positionChanges/0/new", Detail: "unexpected type"}}, errors)
}

func TestFieldOld(t *testing.T) {
	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"val"`), Old: json.RawMessage(`"hey"`)},
	}

	old := "hey"
	v := "not-hey"
	errors := Changes(changes).Field("/thing", &v, Old(&old)).Errors()

	assert.Equal(t, "val", v)
	assert.Empty(t, errors)
}

func TestFieldOldWhenNotMatch(t *testing.T) {
	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"val"`), Old: json.RawMessage(`"not-hey"`)},
	}

	old := "hey"
	v := "not-hey"
	errors := Changes(changes).Field("/thing", &v, Old(&old)).Errors()

	assert.Equal(t, []shared.FieldError{{Source: "/positionChanges/0/old", Detail: "does not match existing value"}}, errors)
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
	errors := Changes(changes).
		Field("/thing", &v, Validate(validate.NotEmpty())).
		Errors()

	assert.Nil(t, errors)
	assert.Equal(t, "val", v)
}

func TestFieldValidateWhenInvalid(t *testing.T) {
	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`""`), Old: jsonNull},
	}

	var v string
	errors := Changes(changes).
		Field("/thing", &v, Validate(validate.NotEmpty())).
		Errors()

	assert.Equal(t, []shared.FieldError{{Source: "/positionChanges/0/new", Detail: "field is required"}}, errors)
}

func TestFieldOldString(t *testing.T) {
	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"new val"`), Old: json.RawMessage(`"old val"`)},
	}

	v := "old val"
	Changes(changes).Field("/thing", &v)

	assert.Equal(t, "new val", v)
}

func TestFieldOldTime(t *testing.T) {
	now := time.Now()
	yesterday := time.Now().Add(-24 * time.Hour)

	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"` + now.Format(time.RFC3339Nano) + `"`), Old: json.RawMessage(`"` + yesterday.Format(time.RFC3339Nano) + `"`)},
	}

	Changes(changes).Field("/thing", &yesterday)

	assert.WithinDuration(t, now, yesterday, time.Second)
}

func TestFieldOldLang(t *testing.T) {
	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"cy"`), Old: json.RawMessage(`"en"`)},
	}

	v := shared.LangEn
	Changes(changes).Field("/thing", &v)

	assert.Equal(t, shared.LangCy, v)
}

func TestFieldOldChannel(t *testing.T) {
	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"online"`), Old: json.RawMessage(`"paper"`)},
	}

	v := shared.ChannelPaper
	Changes(changes).Field("/thing", &v)

	assert.Equal(t, shared.ChannelOnline, v)
}

func TestFieldOldDate(t *testing.T) {
	oldDate, newDate := shared.Date{}, shared.Date{}
	_ = oldDate.UnmarshalText([]byte("2000-11-10"))
	_ = newDate.UnmarshalText([]byte("1990-01-02"))

	changes := []shared.Change{
		{Key: "/thing", New: json.RawMessage(`"1990-01-02"`), Old: json.RawMessage(`"2000-11-10"`)},
	}

	Changes(changes).Field("/thing", &oldDate)
	assert.Equal(t, &newDate, &oldDate)
}

func TestFieldWhenOldDoesNotMatchExisting(t *testing.T) {
	testcases := map[string]json.RawMessage{
		"string": json.RawMessage(`"not same as existing"`),
		"null":   jsonNull,
	}

	for name, oldValue := range testcases {
		t.Run(name, func(t *testing.T) {
			changes := []shared.Change{
				{Key: "/thing", New: json.RawMessage(`"val"`), Old: oldValue},
			}

			v := "existing"
			errors := Changes(changes).Field("/thing", &v).Errors()

			assert.Equal(t, []shared.FieldError{{Source: "/positionChanges/0/old", Detail: "does not match existing value"}}, errors)
		})
	}
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

	assert.Equal(t, []shared.FieldError{{Source: "/positionChanges/0", Detail: "unexpected change provided"}}, errors)
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
	errors := Changes(changes).Each(func(i string, p *Parser) []shared.FieldError {
		switch i {
		case "0":
			p.Field("/thing", &v)
		case "1":
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

	errors := Changes(changes).Each(func(i string, p *Parser) []shared.FieldError {
		var v any
		p.Field("/thing", v)
		return p.Errors()
	}).Errors()

	assert.Equal(t, []shared.FieldError{
		{Source: "/positionChanges/0/key", Detail: "require index"},
		{Source: "/positionChanges/1/key", Detail: "require index"},
	}, errors)
}

func TestEachWhenRequired(t *testing.T) {
	changes := []shared.Change{{Key: "/0/thing", New: json.RawMessage(`"val"`), Old: json.RawMessage(`"old"`)}}

	existing := "old"
	errors := Changes(changes).Each(func(i string, p *Parser) []shared.FieldError {
		p.Field("/thing", &existing)
		return p.Errors()
	}, 1).Errors()

	assert.Equal(t, []shared.FieldError{
		{Source: "/positionChanges", Detail: "missing /1/thing"},
	}, errors)
}

func TestEachWhenOutOfRange(t *testing.T) {
	changes := []shared.Change{
		{Key: "/0/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
		{Key: "/1/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
		{Key: "/2/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	errors := Changes(changes).Each(func(i string, p *Parser) []shared.FieldError {
		idx, _ := strconv.Atoi(i)
		if idx > 0 {
			return p.OutOfRange()
		}

		return p.Errors()
	}).Errors()

	assert.ElementsMatch(t, []shared.FieldError{
		{Source: "/positionChanges/1/key", Detail: "index out of range"},
		{Source: "/positionChanges/2/key", Detail: "index out of range"},
	}, errors)
}

func TestEachWhenNotConsumed(t *testing.T) {
	changes := []shared.Change{
		{Key: "/0/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
		{Key: "/1/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
		{Key: "/2/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	errors := Changes(changes).Each(func(i string, p *Parser) []shared.FieldError {
		idx, _ := strconv.Atoi(i)
		if idx == 0 {
			var v string
			p.Field("/thing", &v)
		}

		return p.Consumed()
	}).Errors()

	assert.ElementsMatch(t, []shared.FieldError{
		{Source: "/positionChanges/1", Detail: "unexpected change provided"},
		{Source: "/positionChanges/2", Detail: "unexpected change provided"},
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
		{Source: "/positionChanges/1", Detail: "unexpected change provided"},
		{Source: "/positionChanges/2", Detail: "unexpected change provided"},
	}, errors)
}

func TestPrefixWhenMissing(t *testing.T) {
	changes := []shared.Change{}

	errors := Changes(changes).Prefix("/a", func(p *Parser) []shared.FieldError { return nil }).Consumed()

	assert.ElementsMatch(t, []shared.FieldError{
		{Source: "/positionChanges", Detail: "missing /a/..."},
	}, errors)
}

func TestPrefixOptional(t *testing.T) {
	changes := []shared.Change{
		{Key: "/a/thing", New: json.RawMessage(`"val"`), Old: jsonNull},
	}

	var v string
	errors := Changes(changes).Prefix("/a", func(p *Parser) []shared.FieldError {
		return p.
			Field("/thing", &v).
			Consumed()
	}, Optional()).Consumed()

	assert.Equal(t, "val", v)
	assert.Empty(t, errors)
}

func TestOptionalPrefixWhenMissing(t *testing.T) {
	changes := []shared.Change{}
	errors := Changes(changes).Prefix("/a", func(p *Parser) []shared.FieldError { return nil }, Optional()).Consumed()

	assert.Empty(t, errors)
}
