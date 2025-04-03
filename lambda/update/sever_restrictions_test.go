package main

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSeverRestrictionsApply(t *testing.T) {

	testcases := map[string]struct {
		newRestriction string
		oldRestriction string
	}{
		"change restrictions": {
			newRestriction: "I do want",
			oldRestriction: "I do not want to x",
		},
		"can blank restrictions": {
			newRestriction: "",
			oldRestriction: "I want to x",
		},
	}

	for scenario, tc := range testcases {
		lpa := &shared.Lpa{
			LpaInit: shared.LpaInit{
				RestrictionsAndConditions: tc.newRestriction,
			},
		}

		s := SeverRestrictions{
			restrictionsAndConditions: tc.oldRestriction,
		}

		t.Run(scenario, func(t *testing.T) {
			errors := s.Apply(lpa)
			assert.Empty(t, errors)
			assert.Equal(t, s.restrictionsAndConditions, lpa.RestrictionsAndConditions)
		})
	}
}

func TestValidateSeverRestrictions(t *testing.T) {

	update := shared.Update{
		Type: "SEVER_RESTRICTIONS_AND_CONDITIONS",
		Changes: []shared.Change{
			{
				Key: "/restrictionsAndConditions",
				New: json.RawMessage(`"I want"`),
				Old: json.RawMessage(`"I do not want"`),
			},
		},
	}

	lpa := &shared.Lpa{
		LpaInit: shared.LpaInit{
			RestrictionsAndConditions: "I do not want",
		},
	}

	_, errors := validateUpdate(update, lpa)
	assert.ElementsMatch(t, errors, errors)
}
