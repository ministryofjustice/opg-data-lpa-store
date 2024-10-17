package shared

import (
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
	IndependentWitness                          *IndependentWitness     `json:"independentWitness,omitempty"`
	AuthorisedSignatory                         *AuthorisedSignatory    `json:"authorisedSignatory,omitempty"`
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
	WitnessedByCertificateProviderAt            time.Time               `json:"witnessedByCertificateProviderAt"`
	WitnessedByIndependentWitnessAt             *time.Time              `json:"witnessedByIndependentWitnessAt,omitempty"`
	CertificateProviderNotRelatedConfirmedAt    *time.Time              `json:"certificateProviderNotRelatedConfirmedAt,omitempty"`
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
	LpaStatusCancelled              = LpaStatus("cancelled")
)

func (l LpaStatus) IsValid() bool {
	return l == LpaStatusInProgress || l == LpaStatusStatutoryWaitingPeriod || l == LpaStatusRegistered || l == LpaStatusCannotRegister || l == LpaStatusWithdrawn || l == LpaStatusCancelled
}
