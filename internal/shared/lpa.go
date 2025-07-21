package shared

import (
	"strconv"
	"time"
)

type LpaInit struct {
	LpaType                                       LpaType                 `json:"lpaType"`
	Channel                                       Channel                 `json:"channel"`
	Language                                      Lang                    `json:"language"`
	Donor                                         Donor                   `json:"donor"`
	Attorneys                                     []Attorney              `json:"attorneys"`
	TrustCorporations                             []TrustCorporation      `json:"trustCorporations,omitempty"`
	CertificateProvider                           CertificateProvider     `json:"certificateProvider"`
	PeopleToNotify                                []PersonToNotify        `json:"peopleToNotify,omitempty"`
	IndependentWitness                            *IndependentWitness     `json:"independentWitness,omitempty"`
	AuthorisedSignatory                           *AuthorisedSignatory    `json:"authorisedSignatory,omitempty"`
	HowAttorneysMakeDecisions                     HowMakeDecisions        `json:"howAttorneysMakeDecisions,omitempty"`
	HowAttorneysMakeDecisionsDetails              string                  `json:"howAttorneysMakeDecisionsDetails,omitempty"`
	HowAttorneysMakeDecisionsIsDefault            bool                    `json:"howAttorneysMakeDecisionsIsDefault,omitempty"`
	HowReplacementAttorneysMakeDecisions          HowMakeDecisions        `json:"howReplacementAttorneysMakeDecisions,omitempty"`
	HowReplacementAttorneysMakeDecisionsDetails   string                  `json:"howReplacementAttorneysMakeDecisionsDetails,omitempty"`
	HowReplacementAttorneysMakeDecisionsIsDefault bool                    `json:"howReplacementAttorneysMakeDecisionsIsDefault,omitempty"`
	HowReplacementAttorneysStepIn                 HowStepIn               `json:"howReplacementAttorneysStepIn,omitempty"`
	HowReplacementAttorneysStepInDetails          string                  `json:"howReplacementAttorneysStepInDetails,omitempty"`
	WhenTheLpaCanBeUsed                           CanUse                  `json:"whenTheLpaCanBeUsed,omitempty"`
	WhenTheLpaCanBeUsedIsDefault                  bool                    `json:"whenTheLpaCanBeUsedIsDefault,omitempty"`
	LifeSustainingTreatmentOption                 LifeSustainingTreatment `json:"lifeSustainingTreatmentOption,omitempty"`
	LifeSustainingTreatmentOptionIsDefault        bool                    `json:"lifeSustainingTreatmentOptionIsDefault,omitempty"`
	RestrictionsAndConditions                     string                  `json:"restrictionsAndConditions,omitempty"`
	RestrictionsAndConditionsImages               []FileUpload            `json:"restrictionsAndConditionsImages,omitempty"`
	SignedAt                                      time.Time               `json:"signedAt"`
	WitnessedByCertificateProviderAt              time.Time               `json:"witnessedByCertificateProviderAt"`
	WitnessedByIndependentWitnessAt               *time.Time              `json:"witnessedByIndependentWitnessAt,omitempty"`
	CertificateProviderNotRelatedConfirmedAt      *time.Time              `json:"certificateProviderNotRelatedConfirmedAt,omitempty"`
}

type Lpa struct {
	LpaInit
	Uid                             string     `json:"uid"`
	Status                          LpaStatus  `json:"status"`
	RegistrationDate                *time.Time `json:"registrationDate,omitempty"`
	UpdatedAt                       time.Time  `json:"updatedAt"`
	RestrictionsAndConditionsImages []File     `json:"restrictionsAndConditionsImages,omitempty"`
	Notes                           []Note     `json:"notes,omitempty"`
}

type Note struct {
	Type     string            `json:"type"`
	Datetime string            `json:"datetime"`
	Values   map[string]string `json:"values"`
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
	LpaStatusDoNotRegister          = LpaStatus("do-not-register")
	LpaStatusExpired                = LpaStatus("expired")
)

func (l LpaStatus) IsValid() bool {
	return l == LpaStatusInProgress || l == LpaStatusStatutoryWaitingPeriod || l == LpaStatusRegistered || l == LpaStatusCannotRegister || l == LpaStatusWithdrawn || l == LpaStatusCancelled || l == LpaStatusDoNotRegister || l == LpaStatusExpired
}

func (lpa *Lpa) FindAttorneyIndex(changeKey string) (int, bool) {
	if idx, err := strconv.Atoi(changeKey); err == nil && idx < len(lpa.Attorneys) {
		return idx, true
	}

	for i, attorney := range lpa.Attorneys {
		if attorney.UID == changeKey {
			return i, true
		}
	}

	return 0, false
}

func (lpa *Lpa) FindTrustCorporationIndex(changeKey string) (int, bool) {
	if idx, err := strconv.Atoi(changeKey); err == nil && idx < len(lpa.TrustCorporations) {
		return idx, true
	}

	for i, tc := range lpa.TrustCorporations {
		if tc.UID == changeKey {
			return i, true
		}
	}

	return 0, false
}

func (lpa *Lpa) AddNote(note Note) {
	lpa.Notes = append(lpa.Notes, note)
}
