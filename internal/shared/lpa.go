package shared

import "time"

type LpaInit struct {
	Donor     Donor      `json:"donor" dynamodbav:""`
	Attorneys []Attorney `json:"attorneys" dynamodbav:""`
}

type Lpa struct {
	LpaInit
	Uid              string    `json:"uid" dynamodbav:""`
	Status           LpaStatus `json:"status" dynamodbav:""`
	RegistrationDate time.Time `json:"registrationDate" dynamodbav:""`
	UpdatedAt        time.Time `json:"updatedAt" dynamodbav:""`
}

type LpaStatus string

const (
	LpaStatusProcessing = LpaStatus("processing")
	LpaStatusRegistered = LpaStatus("registered")
)
