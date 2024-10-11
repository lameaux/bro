package checker_test

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/lameaux/bro/internal/client/checker"
	"github.com/lameaux/bro/internal/client/config"
)

func TestChecker_Validate(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusOK,
	}

	checks := []*config.Check{
		{
			Type:   checker.TypeHTTPCode,
			Equals: "200",
		},
	}

	responseChecker := checker.New(checks)

	results, success := responseChecker.Validate(resp)
	for _, result := range results {
		if !result.Pass {
			t.Errorf("check failed for value %v", result.Actual)
		}
	}

	if !success {
		t.Errorf("validation failed")
	}
}

func TestRunCheck(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusOK,
	}

	tests := []struct {
		name  string
		check *config.Check
		pass  bool
		err   error
	}{
		{
			name:  "invalid type",
			check: &config.Check{Type: "invalid"},
			err:   checker.ErrUnknownCheckType,
		},
		{
			name:  "valid http code",
			check: &config.Check{Type: checker.TypeHTTPCode, Equals: "200"},
			pass:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := checker.RunCheck(tt.check, resp)
			if result.Pass != tt.pass {
				t.Errorf("pass equals %v; expected %v", result.Pass, tt.pass)
			}

			if !errors.Is(result.Error, tt.err) {
				t.Errorf("err equals %v; expected %v", result.Error, tt.err)
			}
		})
	}
}

func TestCheckHTTPCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		expect     int
		pass       bool
	}{
		{
			name:       "valid status",
			statusCode: 200,
			expect:     200,
			pass:       true,
		},
		{
			name:       "invalid status",
			statusCode: 500,
			expect:     200,
			pass:       false,
		},
		{
			name:       "invalid check",
			statusCode: 200,
			pass:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			check := &config.Check{Equals: strconv.Itoa(tt.expect)}
			resp := &http.Response{StatusCode: tt.statusCode}

			result := checker.CheckHTTPCode(check, resp)
			if result.Pass != tt.pass {
				t.Errorf("pass equals %v; expected %v", result.Pass, tt.pass)
			}

			if result.Actual != strconv.Itoa(tt.statusCode) {
				t.Errorf("actual equals %v; expected %d", result.Actual, tt.statusCode)
			}
		})
	}
}

func TestCheckHTTPHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		headerName  string
		headerValue string

		checkName     string
		checkEquals   string
		checkContains string
		checkPass     bool
	}{
		{
			name:        "valid header and value is equal",
			headerName:  "Content-Type",
			headerValue: "text/html",

			checkName:   "Content-Type",
			checkEquals: "text/html",
			checkPass:   true,
		},
		{
			name:        "valid header and value contains",
			headerName:  "Content-Type",
			headerValue: "text/html",

			checkName:     "Content-Type",
			checkContains: "text",
			checkPass:     true,
		},
		{
			name:       "invalid header",
			headerName: "Length",
			checkName:  "Content-Type",
			checkPass:  false,
		},
		{
			name:        "invalid value",
			headerName:  "Content-Type",
			headerValue: "text/plain",

			checkName:   "Content-Type",
			checkEquals: "text/html",
			checkPass:   false,
		},
		{
			name:        "invalid check",
			headerName:  "Content-Type",
			headerValue: "text/plain",

			checkName: "Content-Type",
			checkPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			check := &config.Check{Name: tt.checkName, Equals: tt.checkEquals, Contains: tt.checkContains}
			headers := make(map[string][]string)
			headers[tt.headerName] = []string{tt.headerValue}
			resp := &http.Response{Header: headers}

			result := checker.CheckHTTPHeader(check, resp)
			if result.Pass != tt.checkPass {
				t.Errorf("pass returned %v; expected %v", result.Pass, tt.checkPass)
			}

			if result.Actual != tt.headerValue {
				t.Errorf("actual equals %v; expected %v", result.Actual, tt.headerValue)
			}
		})
	}
}

func TestCheckHTTPBody(t *testing.T) {
	t.Parallel()

	longStr := strings.Repeat("a", 1000)

	tests := []struct {
		name        string
		body        string
		checkEquals string
		checkPass   bool
	}{
		{
			name:        "valid body",
			body:        "hello",
			checkEquals: "hello",
			checkPass:   true,
		},
		{
			name:        "long body",
			body:        longStr,
			checkEquals: longStr,
			checkPass:   true,
		},
		{
			name:        "invalid body",
			body:        "goodbye",
			checkEquals: "hello",
			checkPass:   false,
		},
		{
			name:        "empty body",
			checkEquals: "hello",
			checkPass:   false,
		},
		{
			name:      "invalid check",
			checkPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			check := &config.Check{Equals: tt.checkEquals}
			resp := &http.Response{Body: io.NopCloser(strings.NewReader(tt.body))}

			result := checker.CheckHTTPBody(check, resp)
			if result.Pass != tt.checkPass {
				t.Errorf("pass returned %v; expected %v", result.Pass, tt.checkPass)
			}

			expected := checker.TruncBody(tt.body)
			if result.Actual != expected {
				t.Errorf("actual equals %v; expected %v", result.Actual, expected)
			}
		})
	}
}
