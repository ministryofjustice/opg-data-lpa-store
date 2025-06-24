package shared

import "time"

type Address struct {
	Line1    string `json:"line1"`
	Line2    string `json:"line2,omitempty"`
	Line3    string `json:"line3,omitempty"`
	Town     string `json:"town,omitempty"`
	Postcode string `json:"postcode,omitempty"`
	Country  string `json:"country"`
}

func (a Address) IsZero() bool {
	return a == Address{}
}

type Person struct {
	UID        string `json:"uid"`
	FirstNames string `json:"firstNames"`
	LastName   string `json:"lastName"`
}

type Donor struct {
	Person
	Address                   Address        `json:"address"`
	DateOfBirth               Date           `json:"dateOfBirth"`
	Email                     string         `json:"email"`
	OtherNamesKnownBy         string         `json:"otherNamesKnownBy,omitempty"`
	ContactLanguagePreference Lang           `json:"contactLanguagePreference"`
	IdentityCheck             *IdentityCheck `json:"identityCheck,omitempty"`
}

type CertificateProvider struct {
	Person
	Address                   Address        `json:"address"`
	Email                     string         `json:"email"`
	Phone                     string         `json:"phone"`
	Channel                   Channel        `json:"channel"`
	SignedAt                  *time.Time     `json:"signedAt,omitempty"`
	ContactLanguagePreference Lang           `json:"contactLanguagePreference,omitempty"`
	IdentityCheck             *IdentityCheck `json:"identityCheck,omitempty"`
}

type Channel string

const (
	ChannelOnline = Channel("online")
	ChannelPaper  = Channel("paper")
)

func (e Channel) IsValid() bool {
	return e == ChannelOnline || e == ChannelPaper
}

type AttorneyStatus string

const (
	AttorneyStatusActive   = AttorneyStatus("active")
	AttorneyStatusInactive = AttorneyStatus("inactive")
	AttorneyStatusRemoved  = AttorneyStatus("removed")
)

func (a AttorneyStatus) IsValid() bool {
	return a == AttorneyStatusActive || a == AttorneyStatusRemoved || a == AttorneyStatusInactive
}

type AppointmentType string

const (
	AppointmentTypeOriginal    = AppointmentType("original")
	AppointmentTypeReplacement = AppointmentType("replacement")
)

func (a AppointmentType) IsValid() bool {
	return a == AppointmentTypeOriginal || a == AppointmentTypeReplacement
}

type Attorney struct {
	Person
	Address                   Address         `json:"address"`
	DateOfBirth               Date            `json:"dateOfBirth"`
	Email                     string          `json:"email,omitempty"`
	Status                    AttorneyStatus  `json:"status"`
	AppointmentType           AppointmentType `json:"appointmentType"`
	Mobile                    string          `json:"mobile,omitempty"`
	SignedAt                  *time.Time      `json:"signedAt,omitempty"`
	ContactLanguagePreference Lang            `json:"contactLanguagePreference,omitempty"`
	Channel                   Channel         `json:"channel"`
	CannotMakeJointDecisions  bool            `json:"cannotMakeJointDecisions,omitempty"`
}

type TrustCorporation struct {
	UID                       string          `json:"uid"`
	Name                      string          `json:"name"`
	CompanyNumber             string          `json:"companyNumber"`
	Email                     string          `json:"email,omitempty"`
	AppointmentType           AppointmentType `json:"appointmentType"`
	Address                   Address         `json:"address"`
	Status                    AttorneyStatus  `json:"status"`
	Mobile                    string          `json:"mobile,omitempty"`
	Signatories               []Signatory     `json:"signatories,omitempty"`
	ContactLanguagePreference Lang            `json:"contactLanguagePreference,omitempty"`
	Channel                   Channel         `json:"channel"`
}

type Signatory struct {
	FirstNames        string    `json:"firstNames"`
	LastName          string    `json:"lastName"`
	ProfessionalTitle string    `json:"professionalTitle"`
	SignedAt          time.Time `json:"signedAt"`
}

func (s Signatory) IsZero() bool {
	return s == Signatory{}
}

type PersonToNotify struct {
	Person
	Address Address `json:"address"`
}

type IndependentWitness struct {
	Person
	Phone   string  `json:"phone"`
	Address Address `json:"address"`
}

type AuthorisedSignatory struct {
	Person
}

type IdentityCheck struct {
	CheckedAt time.Time         `json:"checkedAt"`
	Type      IdentityCheckType `json:"type"`
}

type IdentityCheckType string

const (
	IdentityCheckTypeOneLogin   = IdentityCheckType("one-login")
	IdentityCheckTypeOpgPaperId = IdentityCheckType("opg-paper-id")
)

func (e IdentityCheckType) IsValid() bool {
	return e == IdentityCheckTypeOneLogin || e == IdentityCheckTypeOpgPaperId
}

type HowMakeDecisions string

const (
	HowMakeDecisionsUnset                            = HowMakeDecisions("")
	HowMakeDecisionsJointly                          = HowMakeDecisions("jointly")
	HowMakeDecisionsJointlyAndSeverally              = HowMakeDecisions("jointly-and-severally")
	HowMakeDecisionsJointlyForSomeSeverallyForOthers = HowMakeDecisions("jointly-for-some-severally-for-others")
)

func (e HowMakeDecisions) IsValid() bool {
	return e == HowMakeDecisionsJointly || e == HowMakeDecisionsJointlyAndSeverally || e == HowMakeDecisionsJointlyForSomeSeverallyForOthers
}

func (e HowMakeDecisions) Unset() bool {
	return e == HowMakeDecisionsUnset
}

type HowStepIn string

const (
	HowStepInUnset             = HowStepIn("")
	HowStepInAllCanNoLongerAct = HowStepIn("all-can-no-longer-act")
	HowStepInOneCanNoLongerAct = HowStepIn("one-can-no-longer-act")
	HowStepInAnotherWay        = HowStepIn("another-way")
)

func (e HowStepIn) IsValid() bool {
	return e == HowStepInUnset || e == HowStepInAllCanNoLongerAct || e == HowStepInOneCanNoLongerAct || e == HowStepInAnotherWay
}

type CanUse string

const (
	CanUseUnset            = CanUse("")
	CanUseWhenCapacityLost = CanUse("when-capacity-lost")
	CanUseWhenHasCapacity  = CanUse("when-has-capacity")
)

func (e CanUse) IsValid() bool {
	return e == CanUseWhenCapacityLost || e == CanUseWhenHasCapacity
}

func (e CanUse) Unset() bool {
	return e == CanUseUnset
}

type LifeSustainingTreatment string

const (
	LifeSustainingTreatmentUnset   = LifeSustainingTreatment("")
	LifeSustainingTreatmentOptionA = LifeSustainingTreatment("option-a")
	LifeSustainingTreatmentOptionB = LifeSustainingTreatment("option-b")
)

func (e LifeSustainingTreatment) IsValid() bool {
	return e == LifeSustainingTreatmentOptionA || e == LifeSustainingTreatmentOptionB
}

func (e LifeSustainingTreatment) Unset() bool {
	return e == LifeSustainingTreatmentUnset
}
