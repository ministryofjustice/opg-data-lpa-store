package shared

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestGetEventHeader(t *testing.T) {
	values := []string{"a", "b"}

	testcases := map[string]string{
		"titlecase": "My-Header",
		"lowercase": "my-header",
	}

	for scenario, headerName := range testcases {
		t.Run(scenario, func(t *testing.T) {
			event := events.APIGatewayProxyRequest{
				MultiValueHeaders: map[string][]string{
					headerName: values,
				},
			}

			result := GetEventHeader("My-header", event)
			assert.Equal(t, values, result)
		})
	}
}

func TestGetEventHeaderWhenNotFound(t *testing.T) {
	event := events.APIGatewayProxyRequest{
		MultiValueHeaders: map[string][]string{
			"MY-HEADER": []string{"a", "b"},
		},
	}

	result := GetEventHeader("My-header", event)
	assert.Equal(t, []string{}, result)
}
