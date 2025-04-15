package main

import (
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/validate"
	"github.com/ministryofjustice/opg-data-lpa-store/lambda/update/parse"
)

type PaperCertificateProviderAccessOnline struct {
	Email string
}

func (c PaperCertificateProviderAccessOnline) Apply(lpa *shared.Lpa) []shared.FieldError {
	if lpa.CertificateProvider.Channel != shared.ChannelPaper {
		return []shared.FieldError{{Source: "/channel", Detail: "lpa channel is not paper"}}
	}

	lpa.CertificateProvider.Email = c.Email

	return nil
}

func validatePaperCertificateProviderAccessOnline(changes []shared.Change) (Applyable, []shared.FieldError) {
	var data PaperCertificateProviderAccessOnline
	errors := parse.Changes(changes).
		Field("/certificateProvider/email", &data.Email, parse.Validate(validate.NotEmpty())).
		Consumed()

	return data, errors
}
