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

func Run(check config.Check, response *http.Response) (actual string, ok bool, err error) {
	if check.Type == typeHttpCode {
		return checkHttpCode(check, response)
	}

	if check.Type == typeHttpHeader {
		return checkHttpHeader(check, response)
	}

	if check.Type == typeHttpBody {
		return checkHttpBody(check, response)
	}

	return "", false, fmt.Errorf("unknown check type: %v", check.Type)
}

func checkHttpCode(check config.Check, response *http.Response) (actual string, ok bool, err error) {
	actual = strconv.Itoa(response.StatusCode)

	if check.Equals != "" {
		ok = actual == check.Equals
		return
	}

	return
}

func checkHttpHeader(check config.Check, response *http.Response) (actual string, ok bool, err error) {
	actual = response.Header.Get(check.Name)

	if check.Equals != "" {
		ok = actual == check.Equals
		return
	}

	if check.Contains != "" {
		ok = strings.Contains(actual, check.Contains)
		return
	}

	return
}

func checkHttpBody(check config.Check, response *http.Response) (actual string, ok bool, err error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", false, fmt.Errorf("failed to read response body: %w", err)
	}
	bodyString := string(body)

	if len(bodyString) > maxBodyLength {
		actual = bodyString[0:maxBodyLength] + "..."
	} else {
		actual = bodyString
	}

	if check.Equals != "" {
		ok = bodyString == check.Equals
		return
	}

	if check.Contains != "" {
		ok = strings.Contains(bodyString, check.Contains)
		return
	}

	return
}
