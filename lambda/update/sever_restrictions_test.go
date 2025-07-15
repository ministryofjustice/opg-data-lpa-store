package main

import (
	"encoding/json"
	"testing"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestSeverRestrictionsApply(t *testing.T) {
	testcases := map[string]struct {
		newRestriction                   string
		oldRestriction                   string
		updatedRestrictionsAndConditions string
	}{
		"change restrictions": {
			newRestriction:                   "I do want",
			oldRestriction:                   "I do not want to x",
			updatedRestrictionsAndConditions: "I do want",
		},
		"can blank restrictions": {
			newRestriction:                   "",
			oldRestriction:                   "I want to x",
			updatedRestrictionsAndConditions: "All restrictions have been severed from the LPA",
		},
	}

	for scenario, tc := range testcases {
		lpa := &shared.Lpa{
			LpaInit: shared.LpaInit{
				RestrictionsAndConditions: tc.oldRestriction,
			},
			RestrictionsAndConditionsImages: []shared.File{
				{
					Path: "folder/test.png",
					Hash: "fake-hash",
				},
			},
		}

		s := SeverRestrictions{
			restrictionsAndConditions: tc.newRestriction,
		}

		t.Run(scenario, func(t *testing.T) {
			errors := s.Apply(lpa)

			noteValues := lpa.Notes[0]["values"].(map[string]string)

			assert.Empty(t, errors)
			assert.Equal(t, tc.newRestriction, lpa.RestrictionsAndConditions)
			assert.Len(t, lpa.RestrictionsAndConditionsImages, 0)
			assert.Len(t, lpa.Notes, 1)
			assert.Equal(t, "SEVER_RESTRICTIONS_AND_CONDITIONS_V1", lpa.Notes[0]["type"])
			assert.Len(t, noteValues, 1)
			assert.Equal(t, tc.updatedRestrictionsAndConditions, noteValues["updatedRestrictionsAndConditions"])
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
