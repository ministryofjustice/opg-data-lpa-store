package objectstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPutFailureLpaAlreadyExists(t *testing.T) {
	c := New("http://localhost:4566")
	obj, err := c.Get("M-EEEE-QQQQ-TTYY")
	assert.NotEqual(t, nil, obj)
	assert.Equal(t, nil, err)
}

func TestPutSuccess(t *testing.T) {
}

func TestGetFailureLpaDoesNotExist(t *testing.T) {
}

func TestGetSuccess(t *testing.T) {
}
