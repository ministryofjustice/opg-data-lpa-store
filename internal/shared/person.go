package shared

type Address struct {
	Line1    string `json:"line1" dynamodbav:""`
	Line2    string `json:"line2" dynamodbav:""`
	Line3    string `json:"line3" dynamodbav:""`
	Town     string `json:"town" dynamodbav:""`
	Postcode string `json:"postcode" dynamodbav:""`
	Country  string `json:"country" dynamodbav:""`
}

type Person struct {
	FirstNames string  `json:"firstNames" dynamodbav:""`
	LastName   string  `json:"lastName"`
	Address    Address `json:"address" dynamodbav:""`
}

type Donor struct {
	Person
	DateOfBirth       Date   `json:"dateOfBirth" dynamodbav:""`
	Email             string `json:"email" dynamodbav:""`
	OtherNamesKnownBy string `json:"otherNamesKnownBy" dynamodbav:""`
}

type CertificateProvider struct {
	Person
	Email      string     `json:"email" dynamodbav:""`
	CarryOutBy CarryOutBy `json:"carryOutBy"`
}

type CarryOutBy string

const (
	CarryOutByOnline = CarryOutBy("online")
	CarryOutByPaper  = CarryOutBy("paper")
)

func (e CarryOutBy) IsValid() bool {
	return e == CarryOutByOnline || e == CarryOutByPaper
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
	DateOfBirth Date           `json:"dateOfBirth" dynamodbav:""`
	Email       string         `json:"email" dynamodbav:""`
	Status      AttorneyStatus `json:"status" dynamodbav:""`
}

type PersonToNotify struct {
	Person
}

type HowMakeDecisions string

const (
	HowMakeDecisionsUnset                            = HowMakeDecisions("")
	HowMakeDecisionsJointly                          = HowMakeDecisions("jointly")
	HowMakeDecisionsJointlyAndSeverally              = HowMakeDecisions("jointly-and-severally")
	HowMakeDecisionsJointlyForSomeSeverallyForOthers = HowMakeDecisions("mixed")
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
	HowStepInAllCanNoLongerAct = HowStepIn("all")
	HowStepInOneCanNoLongerAct = HowStepIn("one")
	HowStepInAnotherWay        = HowStepIn("other")
)

func (e HowStepIn) IsValid() bool {
	return e == HowStepInUnset || e == HowStepInAllCanNoLongerAct || e == HowStepInOneCanNoLongerAct || e == HowStepInAnotherWay
}

type CanUseWhen string

const (
	CanUseWhenUnset        = CanUseWhen("")
	CanUseWhenCapacityLost = CanUseWhen("when-capacity-lost")
	CanUseWhenHasCapacity  = CanUseWhen("when-has-capacity")
)

func (e CanUseWhen) IsValid() bool {
	return e == CanUseWhenCapacityLost || e == CanUseWhenHasCapacity
}

func (e CanUseWhen) Unset() bool {
	return e == CanUseWhenUnset
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
