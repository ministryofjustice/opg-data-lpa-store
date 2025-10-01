package shared

import "time"

type DonorCorrection struct {
	FirstNames        string
	LastName          string
	OtherNamesKnownBy string
	DateOfBirth       Date
	Address           Address
	Email             string
}

type CertificateProviderCorrection struct {
	FirstNames string
	LastName   string
	Address    Address
	Email      string
	Phone      string
	SignedAt   time.Time
}

type AttorneyCorrection struct {
	Index                    *int
	FirstNames               string
	LastName                 string
	DateOfBirth              Date
	Address                  Address
	Email                    string
	Mobile                   string
	SignedAt                 time.Time
	CannotMakeJointDecisions bool
	AppointmentType          AppointmentType
}

type TrustCorporationCorrection struct {
	Index         *int
	Name          string
	CompanyNumber string
	Email         string
	Address       Address
	Mobile        string
	Signatories   []Signatory
}

type AttorneyAppointmentTypeCorrection struct {
	HowAttorneysMakeDecisions                   HowMakeDecisions
	HowAttorneysMakeDecisionsDetails            string
	HowReplacementAttorneysStepIn               HowStepIn
	HowReplacementAttorneysStepInDetails        string
	HowReplacementAttorneysMakeDecisions        HowMakeDecisions
	HowReplacementAttorneysMakeDecisionsDetails string
	LifeSustainingTreatmentOption               LifeSustainingTreatment
	WhenTheLpaCanBeUsed                         CanUse
}
