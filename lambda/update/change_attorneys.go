package main

import (
	"strconv"

	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
)

type ChangeAttorney struct {
	ChangeAttorneyStatus []ChangeAttorneyStatus
}

type ChangeAttorneyStatus struct {
	Index  *int
	Status shared.AttorneyStatus
}

func (a ChangeAttorney) Apply(lpa *shared.Lpa) []shared.FieldError {
	for _, changeAttorneyStatus := range a.ChangeAttorneyStatus {
		source := "/attorneys/" + strconv.Itoa(*changeAttorneyStatus.Index) + "/status"

		if changeAttorneyStatus.Status == shared.AttorneyStatusInactive && lpa.Attorneys[*changeAttorneyStatus.Index].Status == shared.AttorneyStatusActive {
			return []shared.FieldError{{Source: source, Detail: "An active attorney cannot be made inactive"}}
		}

		if changeAttorneyStatus.Status == shared.AttorneyStatusActive && lpa.Attorneys[*changeAttorneyStatus.Index].Status == shared.AttorneyStatusRemoved {
			return []shared.FieldError{{Source: source, Detail: "A removed attorney cannot be made active"}}
		}

		lpa.Attorneys[*changeAttorneyStatus.Index].Status = changeAttorneyStatus.Status
	}

	return nil
}

func validateChangeAttorney(changes []shared.Change, lpa *shared.Lpa) (ChangeAttorney, []shared.FieldError) {
	var data ChangeAttorney
	key := -1

	errors := parse.Changes(changes).
		Prefix("/attorneys", func(p *parse.Parser) []shared.FieldError {
			return p.
				Each(func(i string, p *parse.Parser) []shared.FieldError {
					var attorneyIdx int

					if len(p.UidChanges) == 0 {
						i, err := strconv.Atoi(i)
						if err != nil {
							return p.OutOfRange()
						}

						attorneyIdx = i
					} else {
						var ok bool
						attorneyIdx, ok = lpa.FindAttorneyIndex(i)
						if !ok {
							return p.OutOfRange()
						}
					}

					key++
					data.ChangeAttorneyStatus = append(data.ChangeAttorneyStatus, ChangeAttorneyStatus{Index: &attorneyIdx, Status: lpa.Attorneys[attorneyIdx].Status})

					return p.
						Field("/status", &data.ChangeAttorneyStatus[key].Status, parse.Validate(validate.Valid())).
						Consumed()
				}).
				Consumed()
		}).
		Consumed()

	return data, errors
}
