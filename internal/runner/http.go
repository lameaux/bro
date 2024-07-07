package runner

import (
	"github.com/Lameaux/bro/internal/config"
	"net/http"
)

func NewHttpClient(reqConfig config.HttpRequest) *http.Client {
	tr := &http.Transport{
		DisableKeepAlives: reqConfig.DisableKeepAlive, // Disable keep-alive
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   reqConfig.Timeout,
	}

	return client
}
