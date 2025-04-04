package parse

import (
	"encoding/json"
	"fmt"
	"regexp"
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
	return fmt.Sprintf("/positionChanges/%d%s", p.pos, after)
}

type changeWithUID struct {
	shared.Change
	uid string
	pos int
}

func (p changeWithUID) Source(after string) string {
	return fmt.Sprintf("/uidChanges/%d%s", p.pos, after)
}

type Parser struct {
	root            string
	positionChanges []changeWithPosition
	UidChanges      []changeWithUID
	errors          []shared.FieldError
}

// Changes constructs a new [Parser] for a set of positionChanges.
func Changes(changes []shared.Change) *Parser {
	parser := &Parser{}

	if len(changes) > 0 {
		uuidPattern := `/[^/]+/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})/`
		re := regexp.MustCompile(uuidPattern)
		match := re.FindStringSubmatch(changes[0].Key)
		foundUUID := ""

		if len(match) > 1 {
			foundUUID = match[1]
		}

		for i, change := range changes {
			if foundUUID != "" {
				parser.UidChanges = append(parser.UidChanges, changeWithUID{Change: change, pos: i, uid: foundUUID})
			} else {
				parser.positionChanges = append(parser.positionChanges, changeWithPosition{Change: change, pos: i})
			}
		}
	}

	return parser
}

// Consumed checks the [Parser] has used all of the positionChanges. It adds an error for any unparsed positionChanges.
func (p *Parser) Consumed() []shared.FieldError {
	for _, change := range p.positionChanges {
		p.errors = append(p.errors, shared.FieldError{Source: change.Source(""), Detail: "unexpected change provided"})
	}

	for _, change := range p.UidChanges {
		p.errors = append(p.errors, shared.FieldError{Source: change.Source(""), Detail: "unexpected change provided"})
	}

	return p.errors
}

// OutOfRange can be used with [Parser.Each] when the index is not in an expected range. It adds an out of range error for all positionChanges.
func (p *Parser) OutOfRange() []shared.FieldError {
	for _, change := range p.positionChanges {
		p.errors = append(p.errors, shared.FieldError{Source: change.Source("/key"), Detail: "index out of range"})
	}

	for _, change := range p.UidChanges {
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
	old       any
	optional  bool
	validator validate.Validator
}

// Old provides the value to use when verifying the correct "old" value is
// provided. The type of v must match the type of existing given to field. This
// option is only needed when you want to track whether a change has been
// provided.
func Old(v any) Option {
	return func(f fieldOpts) fieldOpts {
		f.old = v
		return f
	}
}

// Optional stops [Parser.Field] or [Parser.Prefix] from adding an error when
// the expected key is missing.
func Optional() Option {
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

	for i, change := range p.positionChanges {
		if change.Key == key {
			var old any
			if err := json.Unmarshal(change.Old, &old); err != nil {
				p.errors = append(p.errors, shared.FieldError{Source: change.Source("/old"), Detail: "error marshalling old value"})
			}

			compare := existing
			if options.old != nil {
				compare = options.old
			}

			if !oldEqualsExisting(old, compare) {
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

			p.positionChanges = slices.Delete(p.positionChanges, i, i+1)
			return p
		}
	}

	for i, change := range p.UidChanges {
		if change.Key == key {
			var old any
			if err := json.Unmarshal(change.Old, &old); err != nil {
				p.errors = append(p.errors, shared.FieldError{Source: change.Source("/old"), Detail: "error marshalling old value"})
			}

			compare := existing
			if options.old != nil {
				compare = options.old
			}

			if !oldEqualsExisting(old, compare) {
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

			p.UidChanges = slices.Delete(p.UidChanges, i, i+1)
			return p
		}
	}

	if !options.optional {
		p.errors = append(p.errors, shared.FieldError{Source: "/positionChanges", Detail: "missing " + p.root + key})
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

	case *shared.AttorneyStatus:
		if old == nil {
			return *v == ""
		}

		return shared.AttorneyStatus(old.(string)) == *v

	case *shared.HowMakeDecisions:
		if old == nil {
			return *v == ""
		}

		return shared.HowMakeDecisions(old.(string)) == *v

	case *shared.HowStepIn:
		if old == nil {
			return *v == ""
		}

		return shared.HowStepIn(old.(string)) == *v

	case *shared.CanUse:
		if old == nil {
			return *v == ""
		}

		return shared.CanUse(old.(string)) == *v

	case *shared.LifeSustainingTreatment:
		if old == nil {
			return *v == ""
		}

		return shared.LifeSustainingTreatment(old.(string)) == *v

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
func (p *Parser) Each(fn func(string, *Parser) []shared.FieldError, required ...int) *Parser {
	if len(p.positionChanges) > 0 {
		indexedChanges := map[int][]changeWithPosition{}

		for _, idx := range required {
			indexedChanges[idx] = []changeWithPosition{}
		}

		for _, change := range p.positionChanges {
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

		// because we should be going through all the positionChanges, or they 'require index' so are not valid to use
		p.positionChanges = []changeWithPosition{}

		// so we always run through in a consistent order
		indexes := make([]int, 0, len(indexedChanges))
		for k := range indexedChanges {
			indexes = append(indexes, k)
		}
		slices.Sort(indexes)

		for _, idx := range indexes {
			changes := indexedChanges[idx]
			subParser := &Parser{root: p.root + "/" + strconv.Itoa(idx), positionChanges: changes}
			fn(strconv.Itoa(idx), subParser)
			p.errors = append(p.errors, subParser.errors...)
		}

		return p
	}

	uidChanges := map[string][]changeWithUID{}

	for _, change := range p.UidChanges {
		parts := strings.SplitN(change.Key, "/", 3)
		if len(parts) != 3 || parts[0] != "" {
			p.errors = append(p.errors, shared.FieldError{Source: change.Source("/key"), Detail: "require index"})
			continue
		}

		if parts[1] == "" {
			p.errors = append(p.errors, shared.FieldError{Source: change.Source("/key"), Detail: "require uid"})
			continue
		}

		uidChanges[parts[1]] = append(uidChanges[parts[1]], changeWithUID{
			Change: shared.Change{Key: "/" + parts[2], Old: change.Old, New: change.New},
			pos:    change.pos,
			uid:    change.uid,
		})
	}

	// because we should be going through all the uidChanges, or they 'require index' so are not valid to use
	p.UidChanges = []changeWithUID{}

	// so we always run through in a consistent order
	uids := make([]string, 0, len(uidChanges))
	for k := range uidChanges {
		uids = append(uids, k)
	}
	slices.Sort(uids)

	for _, uid := range uids {
		changes := uidChanges[uid]
		subParser := &Parser{root: p.root + "/" + uid, UidChanges: changes}
		fn(uid, subParser)
		p.errors = append(p.errors, subParser.errors...)
	}

	return p
}

// Prefix will run fn with a [Parser] of any positionChanges with the specified prefix. It
// will add an error if the prefix does not exist.
//
// Consider the positionChanges:
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
	var indexMatching, indexRemaining []changeWithPosition
	var uidMatching, uidRemaining []changeWithUID

	options := fieldOpts{}
	for _, opt := range opts {
		options = opt(options)
	}

	subParser := &Parser{root: p.root + prefix}

	if len(p.positionChanges) > 0 {
		for _, change := range p.positionChanges {
			if strings.HasPrefix(change.Key, prefix+"/") {
				indexMatching = append(indexMatching, changeWithPosition{
					Change: shared.Change{Key: change.Key[len(prefix):], Old: change.Old, New: change.New},
					pos:    change.pos,
				})
			} else {
				indexRemaining = append(indexRemaining, change)
			}
		}

		p.positionChanges = indexRemaining

		if len(indexMatching) == 0 {
			if !options.optional {
				p.errors = append(p.errors, shared.FieldError{Source: "/positionChanges", Detail: "missing " + p.root + prefix + "/..."})
			}
		} else {
			subParser.positionChanges = indexMatching
		}
	} else if len(p.UidChanges) > 0 {
		for _, change := range p.UidChanges {
			if strings.HasPrefix(change.Key, prefix+"/") {
				uidMatching = append(uidMatching, changeWithUID{
					Change: shared.Change{Key: change.Key[len(prefix):], Old: change.Old, New: change.New},
					uid:    change.uid,
					pos:    change.pos,
				})
			} else {
				uidRemaining = append(uidRemaining, change)
			}
		}

		p.UidChanges = uidRemaining

		if len(uidMatching) == 0 {
			if !options.optional {
				p.errors = append(p.errors, shared.FieldError{Source: "/uidChanges", Detail: "missing " + p.root + prefix + "/..."})
			}
		} else {
			subParser.UidChanges = uidMatching
		}
	} else {
		if !options.optional {
			p.errors = append(p.errors, shared.FieldError{Source: "/positionChanges", Detail: "missing " + p.root + prefix + "/..."})
		}
	}

	fn(subParser)
	p.errors = append(p.errors, subParser.errors...)

	return p
}
