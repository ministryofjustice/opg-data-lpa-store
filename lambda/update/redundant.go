package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

func redundantChangeErrors(changes []shared.Change) ([]shared.FieldError, error) {
	if len(changes) == 0 {
		return nil, nil
	}

	var errors []shared.FieldError
	for i, change := range changes {
		redundant, err := isRedundantChange(change.Old, change.New)
		if err != nil {
			return nil, fmt.Errorf("change %d: %w", i, err)
		}

		if redundant {
			errors = append(errors, shared.FieldError{
				Source: fmt.Sprintf("/changes/%d", i),
				Detail: fmt.Sprintf("redundant change for %s", change.Key),
			})
		}
	}

	return errors, nil
}

func isRedundantChange(oldRaw, newRaw json.RawMessage) (bool, error) {
	oldValue, err := decodeJSONValue(oldRaw)
	if err != nil {
		return false, err
	}

	newValue, err := decodeJSONValue(newRaw)
	if err != nil {
		return false, err
	}

	if isNilOrEmpty(oldValue) && isNilOrEmpty(newValue) {
		return true, nil
	}

	return reflect.DeepEqual(oldValue, newValue), nil
}

func decodeJSONValue(raw json.RawMessage) (any, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}

	return value, nil
}

func isNilOrEmpty(value any) bool {
	if value == nil {
		return true
	}

	str, ok := value.(string)
	return ok && str == ""
}
