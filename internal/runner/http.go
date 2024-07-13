package runner

import (
	"github.com/Lameaux/bro/internal/config"
	"github.com/rs/zerolog/log"
	"net/http"
)

func NewHttpClient(reqConfig config.HttpClient) *http.Client {
	log.Debug().
		Bool("disableKeepAlive", reqConfig.DisableKeepAlive).
		Dur("timeout", reqConfig.Timeout).
		Msg("creating http client")

	tr := &http.Transport{
		DisableKeepAlives: reqConfig.DisableKeepAlive,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   reqConfig.Timeout,
	}

	return client
}
