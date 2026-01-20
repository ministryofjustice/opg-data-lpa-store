package main

import (
	"time"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

type Register struct{}

func (r Register) Apply(lpa *shared.Lpa) []shared.FieldError {
	if lpa.Status != shared.LpaStatusStatutoryWaitingPeriod {
		return []shared.FieldError{{Source: "/type", Detail: "status must be statutory-waiting-period to register"}}
	}

	lpa.RegistrationDate = time.Now().UTC()
	lpa.Status = shared.LpaStatusRegistered

	return nil
}

func validateRegister(changes []shared.Change) (Register, []shared.FieldError) {
	if len(changes) > 0 {
		return Register{}, []shared.FieldError{{Source: "/changes", Detail: "expected empty"}}
	}

	return Register{}, nil
}
