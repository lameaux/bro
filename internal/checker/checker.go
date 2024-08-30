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

	return "", true, nil
}

func checkHttpCode(check config.Check, response *http.Response) (actual string, ok bool, err error) {
	actual = strconv.Itoa(response.StatusCode)
	ok = actual == check.Equals

	return
}

func checkHttpHeader(check config.Check, response *http.Response) (actual string, ok bool, err error) {
	actual = response.Header.Get(check.Name)
	ok = actual == check.Equals

	return
}

func checkHttpBody(check config.Check, response *http.Response) (actual string, ok bool, err error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", false, fmt.Errorf("failed to read response body: %w", err)
	}
	bodyString := string(body)

	actual = bodyString[0:100] + "..."
	ok = strings.Contains(bodyString, check.Contains)

	return
}
