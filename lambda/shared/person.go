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
	FirstNames  string  `json:"firstNames" dynamodbav:""`
	Surname     string  `json:"surname" dynamodbav:""`
	DateOfBirth Date    `json:"dateOfBirth" dynamodbav:""`
	Email       string  `json:"email" dynamodbav:""`
	Address     Address `json:"address" dynamodbav:""`
}

type Donor struct {
	Person
	OtherNames string `json:"otherNames" dynamodbav:""`
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
	Status AttorneyStatus `json:"status" dynamodbav:""`
}
