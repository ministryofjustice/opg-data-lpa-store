package shared

import (
	"slices"
	"time"
)

type LpaInit struct {
	LpaType                                     LpaType                 `json:"lpaType"`
	Channel                                     Channel                 `json:"channel"`
	Donor                                       Donor                   `json:"donor"`
	Attorneys                                   []Attorney              `json:"attorneys"`
	TrustCorporations                           []TrustCorporation      `json:"trustCorporations,omitempty"`
	CertificateProvider                         CertificateProvider     `json:"certificateProvider"`
	PeopleToNotify                              []PersonToNotify        `json:"peopleToNotify,omitempty"`
	HowAttorneysMakeDecisions                   HowMakeDecisions        `json:"howAttorneysMakeDecisions,omitempty"`
	HowAttorneysMakeDecisionsDetails            string                  `json:"howAttorneysMakeDecisionsDetails,omitempty"`
	HowReplacementAttorneysMakeDecisions        HowMakeDecisions        `json:"howReplacementAttorneysMakeDecisions,omitempty"`
	HowReplacementAttorneysMakeDecisionsDetails string                  `json:"howReplacementAttorneysMakeDecisionsDetails,omitempty"`
	HowReplacementAttorneysStepIn               HowStepIn               `json:"howReplacementAttorneysStepIn,omitempty"`
	HowReplacementAttorneysStepInDetails        string                  `json:"howReplacementAttorneysStepInDetails,omitempty"`
	WhenTheLpaCanBeUsed                         CanUse                  `json:"whenTheLpaCanBeUsed,omitempty"`
	LifeSustainingTreatmentOption               LifeSustainingTreatment `json:"lifeSustainingTreatmentOption,omitempty"`
	RestrictionsAndConditions                   string                  `json:"restrictionsAndConditions,omitempty"`
	RestrictionsAndConditionsImages             []FileUpload            `json:"restrictionsAndConditionsImages,omitempty"`
	SignedAt                                    time.Time               `json:"signedAt"`
	CertificateProviderNotRelatedConfirmedAt    *time.Time              `json:"certificateProviderNotRelatedConfirmedAt,omitempty"`
}

func (l *Lpa) GetAttorney(uid string) (Attorney, bool) {
	idx := slices.IndexFunc(l.Attorneys, func(a Attorney) bool { return a.UID == uid })
	if idx == -1 {
		return Attorney{}, false
	}

	return l.Attorneys[idx], true
}

func (l *Lpa) PutAttorney(attorney Attorney) {
	idx := slices.IndexFunc(l.Attorneys, func(a Attorney) bool { return a.UID == attorney.UID })
	if idx == -1 {
		l.Attorneys = append(l.Attorneys, attorney)
	} else {
		l.Attorneys[idx] = attorney
	}
}

func (l *Lpa) ActiveAttorneys() (attorneys []Attorney) {
	for _, a := range l.Attorneys {
		if a.Status == AttorneyStatusActive {
			attorneys = append(attorneys, a)
		}
	}

	return attorneys
}

type Lpa struct {
	LpaInit
	Uid                             string     `json:"uid"`
	Status                          LpaStatus  `json:"status"`
	RegistrationDate                *time.Time `json:"registrationDate,omitempty"`
	UpdatedAt                       time.Time  `json:"updatedAt"`
	RestrictionsAndConditionsImages []File     `json:"restrictionsAndConditionsImages,omitempty"`
}

type LpaType string

const (
	LpaTypePersonalWelfare    = LpaType("personal-welfare")
	LpaTypePropertyAndAffairs = LpaType("property-and-affairs")
)

func (e LpaType) IsValid() bool {
	return e == LpaTypePersonalWelfare || e == LpaTypePropertyAndAffairs
}

type LpaStatus string

const (
	LpaStatusInProgress             = LpaStatus("in-progress")
	LpaStatusStatutoryWaitingPeriod = LpaStatus("statutory-waiting-period")
	LpaStatusRegistered             = LpaStatus("registered")
	LpaStatusCannotRegister         = LpaStatus("cannot-register")
	LpaStatusWithdrawn              = LpaStatus("withdrawn")
)
