package checker

import (
	"github.com/lameaux/bro/internal/config"
	"net/http"
	"strconv"
)

const (
	typeHttpCode = "httpCode"
)

func Run(check config.Check, response *http.Response) (actual string, ok bool, err error) {
	if check.Type == typeHttpCode {
		return checkHttpCode(check, response)
	}

	return "", true, nil
}

func checkHttpCode(check config.Check, response *http.Response) (actual string, ok bool, err error) {
	actual = strconv.Itoa(response.StatusCode)
	ok = actual == check.Equals

	return
}
