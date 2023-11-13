package shared

import (
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

func GetEventHeader(headerName string, event events.APIGatewayProxyRequest) []string {
	headerValues, ok := event.MultiValueHeaders[strings.Title(headerName)]
	if !ok {
		headerValues, ok = event.MultiValueHeaders[strings.ToLower(headerName)]
	}

	if !ok {
		return []string{}
	}

	return headerValues
}
