package shared

import (
	"time"
)

type LpaInit struct {
	LpaType                                     LpaType                 `json:"lpaType"`
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
	SignedAt                                    time.Time               `json:"signedAt"`
	CertificateProviderNotRelatedConfirmedAt    *time.Time              `json:"certificateProviderNotRelatedConfirmedAt,omitempty"`
}

type Lpa struct {
	LpaInit
	Uid              string     `json:"uid"`
	Status           LpaStatus  `json:"status"`
	RegistrationDate *time.Time `json:"registrationDate,omitempty"`
	UpdatedAt        time.Time  `json:"updatedAt"`
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
	LpaStatusProcessing     = LpaStatus("processing")
	LpaStatusPerfect        = LpaStatus("perfect")
	LpaStatusRegistered     = LpaStatus("registered")
	LpaStatusCannotRegister = LpaStatus("cannot-register")
)
