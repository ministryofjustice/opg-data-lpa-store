package shared

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

var (
	ProblemInternalServerError = Problem{
		StatusCode: 500,
		Code:       "INTERNAL_SERVER_ERROR",
		Detail:     "Internal server error",
	}
	ProblemInvalidRequest = Problem{
		StatusCode: 400,
		Code:       "INVALID_REQUEST",
		Detail:     "Invalid request",
	}
	ProblemUnauthorisedRequest = Problem{
		StatusCode: 401,
		Code:       "UNAUTHORISED",
		Detail:     "Invalid JWT",
	}
	ProblemNotFoundRequest = Problem{
		StatusCode: 404,
		Code:       "NOT_FOUND",
		Detail:     "Record not found",
	}
)

type Problem struct {
	StatusCode int          `json:"-"`
	Code       string       `json:"code"`
	Detail     string       `json:"detail"`
	Errors     []FieldError `json:"errors,omitempty"`
}

func (problem Problem) Respond() (events.APIGatewayProxyResponse, error) {
	var errorString = ""
	for _, ve := range problem.Errors {
		errorString += fmt.Sprintf("%s %s, ", ve.Source, ve.Detail)
	}

	_ = json.NewEncoder(os.Stdout).Encode(LogEvent{
		ServiceName: "opg-data-lpa-store",
		Level:       slog.LevelInfo.String(),
		Message:     problem.Detail,
		Timestamp:   time.Now(),
		Status:      problem.StatusCode,
		Problem:     problem,
		ErrorString: strings.TrimRight(errorString, ", "),
	})

	code := problem.StatusCode
	body, err := json.Marshal(problem)

	if err != nil {
		code = 500
		body = []byte("{\"code\":\"INTERNAL_SERVER_ERROR\",\"detail\":\"Internal server error\"}")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Body:       string(body),
	}, nil
}

type FieldError struct {
	Source string `json:"source"`
	Detail string `json:"detail"`
}

type LogEvent struct {
	Timestamp   time.Time `json:"time"`
	Level       string    `json:"level"`
	Message     string    `json:"msg"`
	ServiceName string    `json:"service_name"`
	Status      int       `json:"status"`
	Problem     Problem   `json:"problem"`
	ErrorString string    `json:"error_string,omitempty"`
}
