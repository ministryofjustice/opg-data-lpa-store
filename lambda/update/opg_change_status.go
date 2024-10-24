package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
)

type OpgChangeStatus struct {
	Status shared.LpaStatus
}

func (r OpgChangeStatus) Apply(lpa *shared.Lpa) []shared.FieldError {

	if r.Status != shared.LpaStatusCannotRegister && r.Status != shared.LpaStatusCancelled {
		return []shared.FieldError{{Source: "/status", Detail: "Status to be updated should be cannot register or cancelled"}}
	}

	if r.Status == shared.LpaStatusCannotRegister && lpa.Status == shared.LpaStatusRegistered {
		return []shared.FieldError{{Source: "/status", Detail: "Lpa status cannot be registered while changing to cannot register"}}
	}

	if r.Status == shared.LpaStatusCannotRegister && lpa.Status == shared.LpaStatusCancelled {
		return []shared.FieldError{{Source: "/status", Detail: "Lpa status cannot be cancelled while changing to cannot register"}}
	}

	if r.Status == shared.LpaStatusCancelled && lpa.Status != shared.LpaStatusRegistered {
		return []shared.FieldError{{Source: "/status", Detail: "Lpa status has to be registered while changing to cancelled"}}
	}

	lpa.Status = r.Status

	return nil
}

func validateOpgChangeStatus(changes []shared.Change, lpa *shared.Lpa) (OpgChangeStatus, []shared.FieldError) {

	var data OpgChangeStatus

	data.Status = lpa.Status

	errors := parse.Changes(changes).
		Field("/status", &data.Status, parse.Validate(func() []shared.FieldError {
			return validate.IsValid("", data.Status)
		})).
		Consumed()

	return data, errors
}