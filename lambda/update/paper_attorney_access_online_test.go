package main

import (
	"encoding/json"
	"testing"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestPaperAttorneyAccessOnlineApply(t *testing.T) {
	idx := 0

	a := PaperAttorneyAccessOnline{
		Index: &idx,
		Email: "a@example.com",
	}

	lpa := shared.Lpa{
		LpaInit: shared.LpaInit{
			Attorneys: []shared.Attorney{
				{Channel: shared.ChannelPaper},
			},
		},
	}

	errors := a.Apply(&lpa)

	assert.Len(t, errors, 0)
	assert.Equal(t, "a@example.com", lpa.Attorneys[0].Email)
}

func TestPaperAttorneyAccessOnlineApplyWhenNotPaper(t *testing.T) {
	idx := 0

	a := PaperAttorneyAccessOnline{
		Index: &idx,
		Email: "a@example.com",
	}

	lpa := shared.Lpa{
		LpaInit: shared.LpaInit{
			Attorneys: []shared.Attorney{
				{Channel: shared.ChannelOnline},
			},
		},
	}

	errors := a.Apply(&lpa)

	assert.Len(t, errors, 1)
	assert.Equal(t, []shared.FieldError{{Source: "/channel", Detail: "lpa channel is not paper"}}, errors)
}

func TestPaperAttorneyAccessOnlineValidate(t *testing.T) {
	idx := 1

	testcases := map[string]struct {
		update        shared.Update
		expectedApply PaperAttorneyAccessOnline
		expectedError []shared.FieldError
	}{
		"valid uid": {
			update: shared.Update{
				Type: "PAPER_ATTORNEY_ACCESS_ONLINE",
				Changes: []shared.Change{
					{Key: "/attorneys/a-uid/email", Old: jsonNull, New: json.RawMessage(`"a@example.com"`)},
				}},
			expectedApply: PaperAttorneyAccessOnline{
				Index: &idx,
				Email: "a@example.com",
			},
		},
		"valid index": {
			update: shared.Update{
				Type: "PAPER_ATTORNEY_ACCESS_ONLINE",
				Changes: []shared.Change{
					{Key: "/attorneys/1/email", Old: jsonNull, New: json.RawMessage(`"a@example.com"`)},
				}},
			expectedApply: PaperAttorneyAccessOnline{
				Index: &idx,
				Email: "a@example.com",
			},
		},
		"missing email value": {
			update: shared.Update{
				Type: "PAPER_ATTORNEY_ACCESS_ONLINE",
				Changes: []shared.Change{
					{Key: "/attorneys/a-uid/email", Old: jsonNull, New: jsonNull},
				}},
			expectedApply: PaperAttorneyAccessOnline{
				Index: &idx,
			},
			expectedError: []shared.FieldError{
				{Source: "/changes/0/new", Detail: "field is required"},
			},
		},
		"unexpected field": {
			update: shared.Update{
				Type: "PAPER_ATTORNEY_ACCESS_ONLINE",
				Changes: []shared.Change{
					{Key: "/attorneys/a-uid/name", Old: jsonNull, New: json.RawMessage(`"name"`)},
				}},
			expectedApply: PaperAttorneyAccessOnline{
				Index: &idx,
			},
			expectedError: []shared.FieldError{
				{Source: "/changes", Detail: "missing /attorneys/a-uid/email"},
				{Source: "/changes/0", Detail: "unexpected change provided"},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			apply, err := validateUpdate(tc.update, &shared.Lpa{
				LpaInit: shared.LpaInit{
					Attorneys: []shared.Attorney{
						{Channel: shared.ChannelPaper, Person: shared.Person{UID: "another-uid"}},
						{Channel: shared.ChannelPaper, Person: shared.Person{UID: "a-uid"}},
					},
				},
			})

			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedApply, apply)
		})
	}
}
