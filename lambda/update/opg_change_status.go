package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

type OpgChangeStatus struct{}

func (r OpgChangeStatus) Apply(lpa *shared.Lpa) []shared.FieldError {
	if lpa.Status == shared.LpaStatusRegistered {
		return []shared.FieldError{{Source: "/type", Detail: "status must not be registered"}}
	}

	lpa.Status = shared.LpaStatusCannotRegister

	return nil
}

func validateOpgChangeStatus(changes []shared.Change) (OpgChangeStatus, []shared.FieldError) {

	return OpgChangeStatus{}, nil
}
