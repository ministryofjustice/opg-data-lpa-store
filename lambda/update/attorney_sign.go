package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
)

type AttorneySign struct {
	Index                     *int
	Mobile                    string
	SignedAt                  time.Time
	ContactLanguagePreference shared.Lang
}

func (a AttorneySign) Apply(lpa *shared.Lpa) []shared.FieldError {
	if !lpa.Attorneys[*a.Index].SignedAt.IsZero() {
		return []shared.FieldError{{Source: "/type", Detail: "attorney cannot sign again"}}
	}

	lpa.Attorneys[*a.Index].Mobile = a.Mobile
	lpa.Attorneys[*a.Index].SignedAt = a.SignedAt
	lpa.Attorneys[*a.Index].ContactLanguagePreference = a.ContactLanguagePreference

	return nil
}

func validateAttorneySign(changes []shared.Change) (data AttorneySign, errors []shared.FieldError) {
	for i, change := range changes {
		after, ok := strings.CutPrefix(change.Key, "/attorneys/")
		if !ok {
			errors = append(errors, shared.FieldError{Source: fmt.Sprintf("/changes/%d", i), Detail: "change not allowed for type"})
			continue
		}

		if !bytes.Equal(change.Old, []byte("null")) {
			errors = append(errors, shared.FieldError{Source: fmt.Sprintf("/changes/%d/old", i), Detail: "field must be null"})
		}

		parts := strings.SplitN(after, "/", 2)
		attorneyIndex, err := strconv.Atoi(parts[0])
		if err != nil {
			errors = append(errors, shared.FieldError{Source: fmt.Sprintf("/changes/%d", i), Detail: "change not allowed for type"})
			continue
		}
		if data.Index != nil && *data.Index != attorneyIndex {
			errors = append(errors, shared.FieldError{Source: fmt.Sprintf("/changes/%d/key", i), Detail: "must be for same attorney"})
			continue
		} else {
			data.Index = &attorneyIndex
		}

		newKey := fmt.Sprintf("/changes/%d/new", i)
		switch parts[1] {
		case "mobile":
			if err := json.Unmarshal(change.New, &data.Mobile); err != nil {
				errors = errorMustBeString(errors, newKey)
			}
		case "signedAt":
			if err := json.Unmarshal(change.New, &data.SignedAt); err != nil {
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
			errors = errorMissing(errors, fmt.Sprintf("/attorneys/%d/mobile", *data.Index))
		}

		if data.SignedAt.IsZero() {
			errors = errorMissing(errors, fmt.Sprintf("/attorneys/%d/signedAt", *data.Index))
		}

		if data.ContactLanguagePreference == shared.Lang("") {
			errors = errorMissing(errors, fmt.Sprintf("/attorneys/%d/contactLanguagePreference", *data.Index))
		}
	}

	return data, errors
}
