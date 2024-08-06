package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

type DonorWithdrawLpa struct{}

func (d DonorWithdrawLpa) Apply(lpa *shared.Lpa) []shared.FieldError {
	if lpa.Status == shared.LpaStatusWithdrawn {
		return []shared.FieldError{{Source: "/type", Detail: "lpa has already been withdrawn"}}
	}

	if lpa.Status == shared.LpaStatusRegistered {
		return []shared.FieldError{{Source: "/type", Detail: "cannot withdraw a registered"}}
	}

	if lpa.Status == shared.LpaStatusCannotRegister {
		return []shared.FieldError{{Source: "/type", Detail: "cannot withdraw an unregisterable lpa"}}
	}

	lpa.Status = shared.LpaStatusWithdrawn

	return nil
}

func validateDonorWithdrawLPA(changes []shared.Change) (DonorWithdrawLpa, []shared.FieldError) {
	if len(changes) > 0 {
		return DonorWithdrawLpa{}, []shared.FieldError{{Source: "/changes", Detail: "expected empty"}}
	}

	return DonorWithdrawLpa{}, nil
}
