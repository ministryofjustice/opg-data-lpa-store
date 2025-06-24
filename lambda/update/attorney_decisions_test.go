package main

//
//func TestAttorneyDecisionsApply(t *testing.T) {
//
//	testcases := map[string]struct {
//		newDecision bool
//		oldDecision bool
//	}{
//		"change restrictions": {
//			newDecision: true,
//			oldDecision: false,
//		},
//		"can blank restrictions": {
//			newDecision: false,
//			oldDecision: nil,
//		},
//	}
//
//	for scenario, tc := range testcases {
//		lpa := &shared.Lpa{
//			LpaInit: shared.LpaInit{
//				RestrictionsAndConditions: tc.newRestriction,
//			},
//			RestrictionsAndConditionsImages: []shared.File{
//				{
//					Path: "folder/test.png",
//					Hash: "fake-hash",
//				},
//			},
//		}
//
//		s := SeverRestrictions{
//			restrictionsAndConditions: tc.oldRestriction,
//		}
//
//		t.Run(scenario, func(t *testing.T) {
//			errors := s.Apply(lpa)
//			assert.Empty(t, errors)
//			assert.Equal(t, s.restrictionsAndConditions, lpa.RestrictionsAndConditions)
//			assert.Len(t, lpa.RestrictionsAndConditionsImages, 0)
//		})
//	}
//}
//
//func TestValidateSeverRestrictions(t *testing.T) {
//
//	update := shared.Update{
//		Type: "SEVER_RESTRICTIONS_AND_CONDITIONS",
//		Changes: []shared.Change{
//			{
//				Key: "/restrictionsAndConditions",
//				New: json.RawMessage(`"I want"`),
//				Old: json.RawMessage(`"I do not want"`),
//			},
//		},
//	}
//
//	lpa := &shared.Lpa{
//		LpaInit: shared.LpaInit{
//			RestrictionsAndConditions: "I do not want",
//		},
//	}
//
//	_, errors := validateUpdate(update, lpa)
//	assert.ElementsMatch(t, errors, errors)
//}
