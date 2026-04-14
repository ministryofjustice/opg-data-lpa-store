package shared

import (
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetEventHeader(headerName string, event events.APIGatewayProxyRequest) []string {
	titleCaser := cases.Title(language.English)
	titleHeader := titleCaser.String(headerName)

	headerValues, ok := event.MultiValueHeaders[titleHeader]
	if !ok {
		headerValues, ok = event.MultiValueHeaders[strings.ToLower(headerName)]
	}

	if !ok {
		headerValues = []string{}
	}

	return headerValues
}
