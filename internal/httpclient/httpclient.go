package httpclient

import (
	"github.com/Lameaux/bro/internal/config"
	"github.com/rs/zerolog/log"
	"net/http"
)

func New(reqConfig config.HttpClient) *http.Client {
	log.Debug().
		Bool("disableKeepAlive", reqConfig.DisableKeepAlive).
		Dur("timeout", reqConfig.Timeout).
		Msg("creating http client")

	tr := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		DisableKeepAlives:   reqConfig.DisableKeepAlive,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   reqConfig.Timeout,
	}

	return client
}
