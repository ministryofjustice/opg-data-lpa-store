package shared

import "time"

type LpaInit struct {
	LpaType                                     LpaType                 `json:"lpaType"`
	Donor                                       Donor                   `json:"donor"`
	Attorneys                                   []Attorney              `json:"attorneys"`
	CertificateProvider                         CertificateProvider     `json:"certificateProvider"`
	PeopleToNotify                              []PersonToNotify        `json:"peopleToNotify"`
	HowAttorneysMakeDecisions                   HowMakeDecisions        `json:"howAttorneysMakeDecisions"`
	HowAttorneysMakeDecisionsDetails            string                  `json:"howAttorneysMakeDecisionsDetails"`
	HowReplacementAttorneysMakeDecisions        HowMakeDecisions        `json:"howReplacementAttorneysMakeDecisions"`
	HowReplacementAttorneysMakeDecisionsDetails string                  `json:"howReplacementAttorneysMakeDecisionsDetails"`
	HowReplacementAttorneysStepIn               HowStepIn               `json:"howReplacementAttorneysStepIn"`
	HowReplacementAttorneysStepInDetails        string                  `json:"howReplacementAttorneysStepInDetails"`
	WhenTheLpaCanBeUsed                         CanUse                  `json:"whenTheLpaCanBeUsed"`
	LifeSustainingTreatmentOption               LifeSustainingTreatment `json:"lifeSustainingTreatmentOption"`
	Restrictions                                string                  `json:"restrictions"`
	SignedAt                                    time.Time               `json:"signedAt"`
}

type Lpa struct {
	LpaInit
	Uid              string    `json:"uid"`
	Status           LpaStatus `json:"status"`
	RegistrationDate time.Time `json:"registrationDate"`
	UpdatedAt        time.Time `json:"updatedAt"`
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
	LpaStatusProcessing = LpaStatus("processing")
	LpaStatusRegistered = LpaStatus("registered")
)
