package shared

import "time"

type Case struct {
	Uid       string    `json:"uid" dynamodbav:"uid"`
	Version   string    `json:"version" dynamodbav:"version"`
	UpdatedAt time.Time `json:"-" dynamodbav:"updatedAt"`
}
