package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

func redundantChangeErrors(changes []shared.Change) []shared.FieldError {
	if len(changes) == 0 {
		return nil
	}

	var errors []shared.FieldError
	for i, change := range changes {
		redundant, err := isRedundantChange(change.Old, change.New)
		if err != nil {
			continue
		}

		if redundant {
			errors = append(errors, shared.FieldError{
				Source: fmt.Sprintf("/changes/%d", i),
				Detail: fmt.Sprintf("redundant change for %s", change.Key),
			})
		}
	}

	return errors
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

	return reflect.DeepEqual(normalizeComparableValue(oldValue), normalizeComparableValue(newValue)), nil
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

func normalizeComparableValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		normalized := make(map[string]any, len(v))
		for key, item := range v {
			normalized[key] = normalizeComparableValue(item)
		}
		return normalized
	case []any:
		normalized := make([]any, len(v))
		for i, item := range v {
			normalized[i] = normalizeComparableValue(item)
		}
		return normalized
	case json.Number:
		if rat, ok := new(big.Rat).SetString(v.String()); ok {
			return rat.RatString()
		}
		return v.String()
	default:
		return value
	}
}

func isNilOrEmpty(value any) bool {
	if value == nil {
		return true
	}

	str, ok := value.(string)
	return ok && str == ""
}
