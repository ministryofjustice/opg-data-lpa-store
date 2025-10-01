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

type AuthorisedSignatoryCorrection struct {
	FirstNames string
	LastName   string
}

type IndependentWitnessCorrection struct {
	FirstNames string
	LastName   string
	Phone      string
	Address    Address
}
