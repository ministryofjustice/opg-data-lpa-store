package shared

import "time"

type LpaInit struct {
	Type                                        Type                    `json:"type"`
	Donor                                       Donor                   `json:"donor" dynamodbav:""`
	Attorneys                                   []Attorney              `json:"attorneys" dynamodbav:""`
	CertificateProvider                         CertificateProvider     `json:"certificateProvider"`
	PeopleToNotify                              []PersonToNotify        `json:"peopleToNotify"`
	HowAttorneysMakeDecisions                   HowMakeDecisions        `json:"howAttorneysMakeDecisions"`
	HowAttorneysMakeDecisionsDetails            string                  `json:"howAttorneysMakeDecisionsDetails"`
	HowReplacementAttorneysMakeDecisions        HowMakeDecisions        `json:"howReplacementAttorneysMakeDecisions"`
	HowReplacementAttorneysMakeDecisionsDetails string                  `json:"howReplacementAttorneysMakeDecisionsDetails"`
	HowReplacementAttorneysStepIn               HowStepIn               `json:"howReplacementAttorneysStepIn"`
	HowReplacementAttorneysStepInDetails        string                  `json:"howReplacementAttorneysStepInDetails"`
	WhenTheLpaCanBeUsed                         CanUseWhen              `json:"whenTheLpaCanBeUsed"`
	LifeSustainingTreatmentOption               LifeSustainingTreatment `json:"lifeSustainingTreatmentOption"`
	Restrictions                                string                  `json:"restrictions"`
	SignedAt                                    time.Time               `json:"signedAt"`
}

type Lpa struct {
	LpaInit
	Uid              string    `json:"uid" dynamodbav:""`
	Status           LpaStatus `json:"status" dynamodbav:""`
	RegistrationDate time.Time `json:"registrationDate" dynamodbav:""`
	UpdatedAt        time.Time `json:"updatedAt" dynamodbav:""`
}

type Type string

const (
	TypeHealthWelfare   = Type("hw")
	TypePropertyFinance = Type("pfa")
)

func (e Type) IsValid() bool {
	return e == TypeHealthWelfare || e == TypePropertyFinance
}

type LpaStatus string

const (
	LpaStatusProcessing = LpaStatus("processing")
	LpaStatusRegistered = LpaStatus("registered")
)
