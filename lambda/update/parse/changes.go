package parse

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
)

type changeWithPosition struct {
	shared.Change
	pos int
}

func (p changeWithPosition) Source(after string) string {
	return fmt.Sprintf("/changes/%d%s", p.pos, after)
}

type Parser struct {
	root    string
	changes []changeWithPosition
	errors  []shared.FieldError
}

// Changes constructs a new [Parser] for a set of changes.
func Changes(changes []shared.Change) *Parser {
	cs := make([]changeWithPosition, len(changes))
	for i, change := range changes {
		cs[i] = changeWithPosition{Change: change, pos: i}
	}

	return &Parser{changes: cs}
}

// Consumed checks the [Parser] has used all of the changes. It adds an error for any unparsed changes.
func (p *Parser) Consumed() []shared.FieldError {
	for _, change := range p.changes {
		p.errors = append(p.errors, shared.FieldError{Source: change.Source(""), Detail: "unexpected change provided"})
	}

	return p.errors
}

// OutOfRange can be used with [Parser.Each] when the index is not in an expected range. It adds an out of range error for all changes.
func (p *Parser) OutOfRange() []shared.FieldError {
	for _, change := range p.changes {
		p.errors = append(p.errors, shared.FieldError{Source: change.Source("/key"), Detail: "index out of range"})
	}

	return p.errors
}

// Errors returns the current error list for the Parser.
func (p *Parser) Errors() []shared.FieldError {
	return p.errors
}

type Option func(fieldOpts) fieldOpts

type fieldOpts struct {
	optional  bool
	validator validate.Validator
}

// Optional stops [Parser.Field] or [Parser.Prefix] from adding an error when the expected key is missing.
func Optional() func(fieldOpts) fieldOpts {
	return func(f fieldOpts) fieldOpts {
		f.optional = true
		return f
	}
}

func Validate(v validate.Validator) Option {
	return func(f fieldOpts) fieldOpts {
		f.validator = v
		return f
	}
}

// Field will JSON unmarshal the specified key into existing. It will add an error if
// the key does not exist.
//
// Consider the change:
//
//	{"key": "/thing", "old": null, "new": "a string"}
//
// Then to parse to a string s do:
//
//	parser.Field("/thing", &s)
func (p *Parser) Field(key string, existing any, opts ...Option) *Parser {
	options := fieldOpts{}
	for _, opt := range opts {
		options = opt(options)
	}

	for i, change := range p.changes {
		if change.Key == key {
			var old any
			if err := json.Unmarshal(change.Old, &old); err != nil {
				p.errors = append(p.errors, shared.FieldError{Source: change.Source("/old"), Detail: "error marshalling old value"})
			}

			if !oldEqualsExisting(old, existing) {
				p.errors = append(p.errors, shared.FieldError{Source: change.Source("/old"), Detail: "does not match existing value"})
			} else {
				if err := json.Unmarshal(change.New, existing); err != nil {
					p.errors = append(p.errors, shared.FieldError{Source: change.Source("/new"), Detail: "unexpected type"})
				} else if options.validator != nil {
					if msg := options.validator.Valid(existing); msg != "" {
						p.errors = append(p.errors, shared.FieldError{Source: change.Source("/new"), Detail: msg})
					}
				}
			}

			p.changes = slices.Delete(p.changes, i, i+1)
			return p
		}
	}

	if !options.optional {
		p.errors = append(p.errors, shared.FieldError{Source: "/changes", Detail: "missing " + p.root + key})
	}
	return p
}

func oldEqualsExisting(old any, existing any) bool {
	switch v := existing.(type) {
	case *time.Time:
		if old == nil {
			return v.IsZero()
		}

		oldTime, err := time.Parse(time.RFC3339Nano, old.(string))
		if err != nil {
			return false
		}

		return oldTime.Equal(*v)

	case *shared.Lang:
		if old == nil {
			return *v == ""
		}

		return shared.Lang(old.(string)) == *v

	case *shared.Channel:
		if old == nil {
			return *v == ""
		}

		return shared.Channel(old.(string)) == *v

	case *shared.IdentityCheckType:
		if old == nil {
			return *v == ""
		}

		return shared.IdentityCheckType(old.(string)) == *v

	case *shared.LpaStatus:
		if old == nil {
			return *v == ""
		}

		return shared.LpaStatus(old.(string)) == *v

	case *shared.Date:
		if old == nil {
			return v.IsZero()
		}
		oldDate := &shared.Date{}
		_ = oldDate.UnmarshalText([]byte(old.(string)))
		return *oldDate == *v

	case *string:
		if old == nil {
			return *v == ""
		}

		return old.(string) == *v

	case *int:
		if old == nil {
			return *v == 0
		}

		return old.(int) == *v

	default:
		return false
	}
}

// Each will run fn with a [Parser] for any indexed keys. If required is specified
// then those indexes must exist.
//
// Consider the changes:
//
//	{"key": "/0/thing", "old": null, "new": "a string"}
//	{"key": "/1/thing", "old": null, "new": "another string"}
//
// Then to parse to a list of strings s do:
//
//	parser.Each(func(i int, p *Parser) {
//		var v string
//		p.Field("/thing", v)
//		s = append(s, v)
//		return p.Consumed()
//	})
func (p *Parser) Each(fn func(int, *Parser) []shared.FieldError, required ...int) *Parser {
	indexedChanges := map[int][]changeWithPosition{}

	for _, idx := range required {
		indexedChanges[idx] = []changeWithPosition{}
	}

	for _, change := range p.changes {
		parts := strings.SplitN(change.Key, "/", 3)
		if len(parts) != 3 || parts[0] != "" {
			p.errors = append(p.errors, shared.FieldError{Source: change.Source("/key"), Detail: "require index"})
			continue
		}

		idx, err := strconv.Atoi(parts[1])
		if err != nil {
			p.errors = append(p.errors, shared.FieldError{Source: change.Source("/key"), Detail: "require index"})
			continue
		}

		indexedChanges[idx] = append(indexedChanges[idx], changeWithPosition{
			Change: shared.Change{Key: "/" + parts[2], Old: change.Old, New: change.New},
			pos:    change.pos,
		})
	}

	// because we should be going through all the changes, or they 'require index' so are not valid to use
	p.changes = []changeWithPosition{}

	// so we always run through in a consistent order
	indexes := make([]int, 0, len(indexedChanges))
	for k := range indexedChanges {
		indexes = append(indexes, k)
	}
	slices.Sort(indexes)

	for _, idx := range indexes {
		changes := indexedChanges[idx]
		subParser := &Parser{root: p.root + "/" + strconv.Itoa(idx), changes: changes}
		fn(idx, subParser)
		p.errors = append(p.errors, subParser.errors...)
	}

	return p
}

// Prefix will run fn with a [Parser] of any changes with the specified prefix. It
// will add an error if the prefix does not exist.
//
// Consider the changes:
//
//	{"key": "/thing/name", "old": null, "new": "a string"}
//	{"key": "/thing/size", "old": null, "new": 5}
//
// Then to parse to a Thing t do:
//
//	parser.Prefix("/thing", func(p *Parser) {
//		return p.
//			Field("/name", &t.Name).
//			Field("/size", &t.Size).
//			Consumed()
//	})
func (p *Parser) Prefix(prefix string, fn func(*Parser) []shared.FieldError, opts ...Option) *Parser {
	var matching, remaining []changeWithPosition

	options := fieldOpts{}
	for _, opt := range opts {
		options = opt(options)
	}

	for _, change := range p.changes {
		if strings.HasPrefix(change.Key, prefix+"/") {
			matching = append(matching, changeWithPosition{
				Change: shared.Change{Key: change.Key[len(prefix):], Old: change.Old, New: change.New},
				pos:    change.pos,
			})
		} else {
			remaining = append(remaining, change)
		}
	}

	p.changes = remaining

	if len(matching) == 0 {
		if !options.optional {
			p.errors = append(p.errors, shared.FieldError{Source: "/changes", Detail: "missing " + p.root + prefix + "/..."})
		}
	} else {
		subParser := &Parser{root: p.root + prefix, changes: matching}
		fn(subParser)
		p.errors = append(p.errors, subParser.errors...)
	}

	return p
}
