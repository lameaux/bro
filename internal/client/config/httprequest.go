package config

import (
	"io"
	"net/http"
	"strings"
)

type HTTPRequest struct {
	URL       string  `yaml:"url"`
	MethodRaw *string `yaml:"method"`
	BodyRaw   *string `yaml:"body"`
}

func (r *HTTPRequest) Method() string {
	if r.MethodRaw != nil {
		return *r.MethodRaw
	}

	return http.MethodGet
}

func (r *HTTPRequest) BodyReader() io.Reader {
	if r.BodyRaw != nil {
		return strings.NewReader(*r.BodyRaw)
	}

	return nil
}
