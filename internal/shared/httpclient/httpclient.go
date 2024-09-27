package httpclient

import (
	"net/http"

	"github.com/lameaux/bro/internal/client/config"
	"github.com/rs/zerolog/log"
)

const defaultMaxIdleConnsPerHost = 100

func New(conf config.HTTPClient) *http.Client {
	maxIdleConnsPerHost := defaultMaxIdleConnsPerHost
	if conf.MaxIdleConnsPerHost > 0 {
		maxIdleConnsPerHost = conf.MaxIdleConnsPerHost
	}

	log.Debug().
		Bool("disableKeepAlive", conf.DisableKeepAlive).
		Dur("timeout", conf.Timeout).
		Int("maxIdleConnsPerHost", maxIdleConnsPerHost).
		Msg("creating http client")

	transport := &http.Transport{
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		DisableKeepAlives:   conf.DisableKeepAlive,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   conf.Timeout,
	}

	if conf.DisableFollowRedirects {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}
