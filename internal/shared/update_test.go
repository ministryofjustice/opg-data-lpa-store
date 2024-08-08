package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateAuthorUID(t *testing.T) {
	testcases := []string{
		"urn:opg:poas:makeregister:users:123",
		"urn:opg:poas:sirius:users:123",
	}

	for _, urn := range testcases {
		t.Run(urn, func(t *testing.T) {
			uid := Update{Author: urn}.AuthorUID()

			assert.Equal(t, "123", uid)
		})
	}

}

func TestUpdateAuthorUIDWhenInvalidFormat(t *testing.T) {
	testcases := []string{
		"urn:opg:poas:makeregister:not-users:123",
		"urn:opg:poas:makeregister:users:",
	}

	for _, urn := range testcases {
		t.Run(urn, func(t *testing.T) {
			uid := Update{Author: urn}.AuthorUID()

			assert.Equal(t, "", uid)
		})
	}
}
