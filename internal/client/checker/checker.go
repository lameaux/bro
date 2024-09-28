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

	typeHTTPCode   = "httpCode"
	typeHTTPHeader = "httpHeader"
	typeHTTPBody   = "httpBody"
)

var errUnknownCheckType = errors.New("unknown check type")

type CheckResult struct {
	Actual string
	Pass   bool
	Error  error
}

func Run(checks []*config.Check, response *http.Response) ([]CheckResult, bool) {
	results := make([]CheckResult, len(checks))

	success := true

	for i, check := range checks {
		result := runCheck(check, response)
		results[i] = result

		if !result.Pass {
			success = false
		}
	}

	return results, success
}

func runCheck(check *config.Check, response *http.Response) CheckResult {
	if check.Type == typeHTTPCode {
		return checkHTTPCode(check, response)
	}

	if check.Type == typeHTTPHeader {
		return checkHTTPHeader(check, response)
	}

	if check.Type == typeHTTPBody {
		return checkHTTPBody(check, response)
	}

	return CheckResult{
		Error: errUnknownCheckType,
	}
}

func checkHTTPCode(check *config.Check, response *http.Response) CheckResult {
	var result CheckResult

	result.Actual = strconv.Itoa(response.StatusCode)

	if check.Equals != "" {
		result.Pass = result.Actual == check.Equals

		return result
	}

	return result
}

func checkHTTPHeader(check *config.Check, response *http.Response) CheckResult {
	var result CheckResult

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

func checkHTTPBody(check *config.Check, response *http.Response) CheckResult {
	var result CheckResult

	body, err := io.ReadAll(response.Body)
	if err != nil {
		result.Error = fmt.Errorf("failed to read response body: %w", err)

		return result
	}

	bodyString := string(body)

	if len(bodyString) > maxBodyLength {
		result.Actual = bodyString[0:maxBodyLength] + "..."
	} else {
		result.Actual = bodyString
	}

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
