package checker

import (
	"fmt"
	"github.com/lameaux/bro/internal/config"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	maxBodyLength = 100

	typeHttpCode   = "httpCode"
	typeHttpHeader = "httpHeader"
	typeHttpBody   = "httpBody"
)

type CheckResult struct {
	Actual string
	Pass   bool
	Error  error
}

func Run(checks []config.Check, response *http.Response) ([]CheckResult, bool) {
	var results []CheckResult

	success := true
	for _, check := range checks {
		result := runCheck(check, response)
		results = append(results, result)

		if !result.Pass {
			success = false
		}

	}

	return results, success
}

func runCheck(check config.Check, response *http.Response) CheckResult {
	if check.Type == typeHttpCode {
		return checkHttpCode(check, response)
	}

	if check.Type == typeHttpHeader {
		return checkHttpHeader(check, response)
	}

	if check.Type == typeHttpBody {
		return checkHttpBody(check, response)
	}

	return CheckResult{
		Error: fmt.Errorf("unknown check type: %v", check.Type),
	}
}

func checkHttpCode(check config.Check, response *http.Response) CheckResult {
	var result CheckResult

	result.Actual = strconv.Itoa(response.StatusCode)

	if check.Equals != "" {
		result.Pass = result.Actual == check.Equals
		return result
	}

	return result
}

func checkHttpHeader(check config.Check, response *http.Response) CheckResult {
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

func checkHttpBody(check config.Check, response *http.Response) CheckResult {
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
