package shared

func CountAttorneys(as []Attorney, ts []TrustCorporation) (actives, replacements int) {
	for _, a := range as {
		if a.Status == AttorneyStatusActive {
			actives++
		} else if a.Status == AttorneyStatusInactive && a.AppointmentType == AppointmentTypeReplacement {
			replacements++
		}
	}

	for _, t := range ts {
		if t.Status == AttorneyStatusActive {
			actives++
		} else if t.Status == AttorneyStatusInactive && t.AppointmentType == AppointmentTypeReplacement {
			replacements++
		}
	}

	return actives, replacements
}
