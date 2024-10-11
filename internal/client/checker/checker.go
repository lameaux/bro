package checker

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/lameaux/bro/internal/client/config"
)

const (
	maxBodyLength = 100

	TypeHTTPCode   = "httpCode"
	TypeHTTPHeader = "httpHeader"
	TypeHTTPBody   = "httpBody"
)

var ErrUnknownCheckType = errors.New("unknown check type")

type Result struct {
	Actual string
	Pass   bool
	Error  error
}

type Checker struct {
	checks []*config.Check
}

func New(checks []*config.Check) *Checker {
	return &Checker{
		checks: checks,
	}
}

func (c *Checker) Validate(response *http.Response) ([]Result, bool) {
	results := make([]Result, len(c.checks))

	success := true

	for i, check := range c.checks {
		result := RunCheck(check, response)
		results[i] = result

		if !result.Pass {
			success = false
		}
	}

	return results, success
}

func RunCheck(check *config.Check, response *http.Response) Result {
	if check.Type == TypeHTTPCode {
		return CheckHTTPCode(check, response)
	}

	if check.Type == TypeHTTPHeader {
		return CheckHTTPHeader(check, response)
	}

	if check.Type == TypeHTTPBody {
		return CheckHTTPBody(check, response)
	}

	return Result{
		Error: ErrUnknownCheckType,
	}
}

func CheckHTTPCode(check *config.Check, response *http.Response) Result {
	var result Result

	result.Actual = strconv.Itoa(response.StatusCode)

	if check.Equals != "" {
		result.Pass = result.Actual == check.Equals

		return result
	}

	return result
}

func CheckHTTPHeader(check *config.Check, response *http.Response) Result {
	var result Result

	result.Actual = response.Header.Get(check.Name)

	if check.Equals != "" {
		result.Pass = result.Actual == check.Equals

		return result
	}

	if check.Contains != "" {
		result.Pass = strings.Contains(result.Actual, check.Contains)

		return result
	}

	return result
}

func CheckHTTPBody(check *config.Check, response *http.Response) Result {
	var result Result

	body, err := io.ReadAll(response.Body)
	if err != nil {
		result.Error = fmt.Errorf("failed to read response body: %w", err)

		return result
	}

	bodyString := string(body)
	result.Actual = TruncBody(bodyString)

	if check.Equals != "" {
		result.Pass = bodyString == check.Equals

		return result
	}

	if check.Contains != "" {
		result.Pass = strings.Contains(bodyString, check.Contains)

		return result
	}

	return result
}

func TruncBody(s string) string {
	if len(s) > maxBodyLength {
		return s[0:maxBodyLength] + "..."
	}

	return s
}
