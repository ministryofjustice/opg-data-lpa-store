package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestTrustCorporationOptOutApply(t *testing.T) {
	now := time.Now()

	testcases := map[string]struct {
		lpa         *shared.Lpa
		expectedLpa *shared.Lpa
		errors      []shared.FieldError
	}{
		"successful apply": {
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					TrustCorporations: []shared.TrustCorporation{
						{UID: "a", Status: shared.AttorneyStatusActive},
						{UID: "b", Status: shared.AttorneyStatusActive},
						{UID: "c", Status: shared.AttorneyStatusActive},
					},
				},
			},
			expectedLpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					TrustCorporations: []shared.TrustCorporation{
						{UID: "a", Status: shared.AttorneyStatusActive},
						{UID: "b", Status: shared.AttorneyStatusRemoved},
						{UID: "c", Status: shared.AttorneyStatusActive},
					},
				},
			},
		},
		"successful apply to inactive": {
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					TrustCorporations: []shared.TrustCorporation{
						{UID: "a", Status: shared.AttorneyStatusActive},
						{UID: "b", Status: shared.AttorneyStatusInactive},
						{UID: "c", Status: shared.AttorneyStatusActive},
					},
				},
			},
			expectedLpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					TrustCorporations: []shared.TrustCorporation{
						{UID: "a", Status: shared.AttorneyStatusActive},
						{UID: "b", Status: shared.AttorneyStatusRemoved},
						{UID: "c", Status: shared.AttorneyStatusActive},
					},
				},
			},
		},
		"not found": {
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					TrustCorporations: []shared.TrustCorporation{
						{UID: "a", Status: shared.AttorneyStatusActive},
					},
				},
			},
			expectedLpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					TrustCorporations: []shared.TrustCorporation{
						{UID: "a", Status: shared.AttorneyStatusActive},
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/type", Detail: "trust corporation not found"},
			},
		},
		"already signed": {
			lpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					TrustCorporations: []shared.TrustCorporation{
						{UID: "b", Status: shared.AttorneyStatusActive, Signatories: []shared.Signatory{{SignedAt: now}}},
					},
				},
			},
			expectedLpa: &shared.Lpa{
				Status: shared.LpaStatusInProgress,
				LpaInit: shared.LpaInit{
					TrustCorporations: []shared.TrustCorporation{
						{UID: "b", Status: shared.AttorneyStatusActive, Signatories: []shared.Signatory{{SignedAt: now}}},
					},
				},
			},
			errors: []shared.FieldError{
				{Source: "/type", Detail: "trust corporation cannot opt out after signing"},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			c := TrustCorporationOptOut{trustCorporationUID: "b"}

			errors := c.Apply(tc.lpa)

			assert.Equal(t, tc.errors, errors)
			assert.Equal(t, tc.expectedLpa, tc.lpa)
		})
	}
}

func TestValidateUpdateTrustCorporationOptOut(t *testing.T) {
	testcases := map[string]struct {
		update   shared.Update
		errors   []shared.FieldError
		expected TrustCorporationOptOut
	}{
		"valid": {
			update: shared.Update{
				Author:  "urn:opg:poas:makeregister:users:dc487ebb-b39d-45ed-bb6a-7f950fd355c9",
				Type:    "TRUST_CORPORATION_OPT_OUT",
				Changes: []shared.Change{},
			},
			expected: TrustCorporationOptOut{trustCorporationUID: "dc487ebb-b39d-45ed-bb6a-7f950fd355c9"},
		},
		"with changes": {
			update: shared.Update{
				Author: "urn:opg:poas:makeregister:users:dc487ebb-b39d-45ed-bb6a-7f950fd355c9",
				Type:   "TRUST_CORPORATION_OPT_OUT",
				Changes: []shared.Change{
					{
						Key: "/something/someValue",
						New: json.RawMessage(`"not expected"`),
						Old: jsonNull,
					},
				},
			},
			expected: TrustCorporationOptOut{},
			errors: []shared.FieldError{
				{Source: "/changes", Detail: "expected empty"},
			},
		},
		"author UID not valid": {
			update: shared.Update{
				Author:  "urn:opg:poas:makeregister:users:not-a-uid",
				Type:    "TRUST_CORPORATION_OPT_OUT",
				Changes: []shared.Change{},
			},
			expected: TrustCorporationOptOut{},
			errors: []shared.FieldError{
				{Source: "/author", Detail: "invalid format"},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			data, errors := validateUpdate(tc.update, &shared.Lpa{})
			assert.Equal(t, tc.expected, data)
			assert.ElementsMatch(t, tc.errors, errors)
		})
	}
}
