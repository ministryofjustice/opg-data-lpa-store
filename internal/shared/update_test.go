package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURNDetails(t *testing.T) {
	testcases := map[URN]struct {
		UID     string
		Service string
	}{
		"urn:opg:poas:makeregister:users:123": {
			UID:     "123",
			Service: "makeregister",
		},
		"urn:opg:poas:sirius:users:456": {
			UID:     "456",
			Service: "sirius",
		},
	}

	for urn, tc := range testcases {
		t.Run(string(urn), func(t *testing.T) {
			details := urn.Details()

			assert.Equal(t, tc.UID, details.UID)
			assert.Equal(t, tc.Service, details.Service)
		})
	}

}

func TestURNDetailsWhenURNInvalidFormat(t *testing.T) {
	testcases := []URN{
		"urn:opg:poas:makeregister:users:",
		"urn-opg-poas-makeregister-users-123",
	}

	for _, urn := range testcases {
		t.Run(string(urn), func(t *testing.T) {
			details := urn.Details()

			assert.Equal(t, AuthorDetails{}, details)
		})
	}
}
