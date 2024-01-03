package shared

import "time"

type Address struct {
	Line1    string `json:"line1"`
	Line2    string `json:"line2"`
	Line3    string `json:"line3"`
	Town     string `json:"town"`
	Postcode string `json:"postcode"`
	Country  string `json:"country"`
}

func (a Address) IsZero() bool {
	return a.Line1 == "" && a.Line2 == "" && a.Line3 == "" && a.Town == "" && a.Postcode == "" && a.Country == ""
}

type Person struct {
	FirstNames string  `json:"firstNames"`
	LastName   string  `json:"lastName"`
	Address    Address `json:"address"`
}

type Donor struct {
	Person
	DateOfBirth       Date   `json:"dateOfBirth"`
	Email             string `json:"email"`
	OtherNamesKnownBy string `json:"otherNamesKnownBy"`
}

type CertificateProvider struct {
	Person
	Email                     string    `json:"email"`
	Channel                   Channel   `json:"channel"`
	SignedAt                  time.Time `json:"signedAt,omitempty"`
	ContactLanguagePreference Lang      `json:"contactLanguagePreference,omitempty"`
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
	AttorneyStatusActive      = AttorneyStatus("active")
	AttorneyStatusReplacement = AttorneyStatus("replacement")
	AttorneyStatusRemoved     = AttorneyStatus("removed")
)

func (a AttorneyStatus) IsValid() bool {
	return a == AttorneyStatusActive || a == AttorneyStatusReplacement || a == AttorneyStatusRemoved
}

type Attorney struct {
	Person
	DateOfBirth Date           `json:"dateOfBirth"`
	Email       string         `json:"email"`
	Status      AttorneyStatus `json:"status"`
}

type TrustCorporation struct {
	Name          string         `json:"name"`
	CompanyNumber string         `json:"companyNumber"`
	Email         string         `json:"email"`
	Address       Address        `json:"address"`
	Status        AttorneyStatus `json:"status"`
}

type PersonToNotify struct {
	Person
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
