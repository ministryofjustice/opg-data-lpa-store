package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
)

type TrustCorporationSign struct {
	Index                     *int
	Mobile                    string
	Signatories               [2]shared.Signatory
	ContactLanguagePreference shared.Lang
}

func (a TrustCorporationSign) Apply(lpa *shared.Lpa) []shared.FieldError {
	if signatories := lpa.TrustCorporations[*a.Index].Signatories; len(signatories) > 0 && !signatories[0].SignedAt.IsZero() {
		return []shared.FieldError{{Source: "/type", Detail: "trust corporation cannot sign again"}}
	}

	lpa.TrustCorporations[*a.Index].Mobile = a.Mobile
	if a.Signatories[1].IsZero() {
		lpa.TrustCorporations[*a.Index].Signatories = a.Signatories[:1]
	} else {
		lpa.TrustCorporations[*a.Index].Signatories = a.Signatories[:]
	}
	lpa.TrustCorporations[*a.Index].ContactLanguagePreference = a.ContactLanguagePreference

	return nil
}

func validateTrustCorporationSign(changes []shared.Change) (data TrustCorporationSign, errors []shared.FieldError) {
	for i, change := range changes {
		after, ok := strings.CutPrefix(change.Key, "/trustCorporations/")
		if !ok {
			errors = append(errors, shared.FieldError{Source: fmt.Sprintf("/changes/%d", i), Detail: "change not allowed for type"})
			continue
		}

		if !bytes.Equal(change.Old, []byte("null")) {
			errors = append(errors, shared.FieldError{Source: fmt.Sprintf("/changes/%d/old", i), Detail: "field must be null"})
		}

		parts := strings.SplitN(after, "/", 2)
		trustCorporationIndex, err := strconv.Atoi(parts[0])
		if err != nil {
			errors = append(errors, shared.FieldError{Source: fmt.Sprintf("/changes/%d", i), Detail: "change not allowed for type"})
			continue
		}
		if data.Index != nil && *data.Index != trustCorporationIndex {
			errors = append(errors, shared.FieldError{Source: fmt.Sprintf("/changes/%d/key", i), Detail: "must be for same trust corporation"})
			continue
		} else {
			data.Index = &trustCorporationIndex
		}

		newKey := fmt.Sprintf("/changes/%d/new", i)
		switch parts[1] {
		case "mobile":
			if err := json.Unmarshal(change.New, &data.Mobile); err != nil {
				errors = errorMustBeString(errors, newKey)
			}
		case "signatories/0/firstNames":
			if err := json.Unmarshal(change.New, &data.Signatories[0].FirstNames); err != nil {
				errors = errorMustBeString(errors, newKey)
			}
		case "signatories/0/lastName":
			if err := json.Unmarshal(change.New, &data.Signatories[0].LastName); err != nil {
				errors = errorMustBeString(errors, newKey)
			}
		case "signatories/0/professionalTitle":
			if err := json.Unmarshal(change.New, &data.Signatories[0].ProfessionalTitle); err != nil {
				errors = errorMustBeString(errors, newKey)
			}
		case "signatories/0/signedAt":
			if err := json.Unmarshal(change.New, &data.Signatories[0].SignedAt); err != nil {
				errors = errorMustBeDateTime(errors, newKey)
			}
		case "signatories/1/firstNames":
			if err := json.Unmarshal(change.New, &data.Signatories[1].FirstNames); err != nil {
				errors = errorMustBeString(errors, newKey)
			}
		case "signatories/1/lastName":
			if err := json.Unmarshal(change.New, &data.Signatories[1].LastName); err != nil {
				errors = errorMustBeString(errors, newKey)
			}
		case "signatories/1/professionalTitle":
			if err := json.Unmarshal(change.New, &data.Signatories[1].ProfessionalTitle); err != nil {
				errors = errorMustBeString(errors, newKey)
			}
		case "signatories/1/signedAt":
			if err := json.Unmarshal(change.New, &data.Signatories[1].SignedAt); err != nil {
				errors = errorMustBeDateTime(errors, newKey)
			}
		case "contactLanguagePreference":
			if err := json.Unmarshal(change.New, &data.ContactLanguagePreference); err != nil {
				errors = errorMustBeString(errors, newKey)
			} else {
				errors = append(errors, validate.IsValid(newKey, data.ContactLanguagePreference)...)
			}
		default:
			errors = append(errors, shared.FieldError{Source: fmt.Sprintf("/changes/%d", i), Detail: "change not allowed for type"})
		}
	}

	if data.Index == nil {
		errors = append(errors, shared.FieldError{Source: "/changes", Detail: "must be specified"})
	} else {
		if data.Mobile == "" {
			errors = errorMissing(errors, fmt.Sprintf("/trustCorporations/%d/mobile", *data.Index))
		}

		if data.Signatories[0].IsZero() {
			errors = errorMissing(errors, fmt.Sprintf("/trustCorporations/%d/signatories/0", *data.Index))
		} else {
			errors = errorMissingSignatory(errors, fmt.Sprintf("/trustCorporations/%d/signatories/0", *data.Index), data.Signatories[0])
		}

		if !data.Signatories[1].IsZero() {
			errors = errorMissingSignatory(errors, fmt.Sprintf("/trustCorporations/%d/signatories/1", *data.Index), data.Signatories[1])
		}

		if data.ContactLanguagePreference == shared.Lang("") {
			errors = errorMissing(errors, fmt.Sprintf("/trustCorporations/%d/contactLanguagePreference", *data.Index))
		}
	}

	return data, errors
}

func errorMissingSignatory(errors []shared.FieldError, prefix string, signatory shared.Signatory) []shared.FieldError {
	if signatory.FirstNames == "" {
		errors = errorMissing(errors, fmt.Sprintf("%s/firstNames", prefix))
	}

	if signatory.LastName == "" {
		errors = errorMissing(errors, fmt.Sprintf("%s/lastName", prefix))
	}

	if signatory.ProfessionalTitle == "" {
		errors = errorMissing(errors, fmt.Sprintf("%s/professionalTitle", prefix))
	}

	if signatory.SignedAt.IsZero() {
		errors = errorMissing(errors, fmt.Sprintf("%s/signedAt", prefix))
	}

	return errors
}
