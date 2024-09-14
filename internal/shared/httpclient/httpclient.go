package httpclient

import (
	"github.com/lameaux/bro/internal/client/config"
	"github.com/rs/zerolog/log"
	"net/http"
)

const defaultMaxIdleConnsPerHost = 100

func New(conf config.HttpClient) *http.Client {
	maxIdleConnsPerHost := defaultMaxIdleConnsPerHost
	if conf.MaxIdleConnsPerHost > 0 {
		maxIdleConnsPerHost = conf.MaxIdleConnsPerHost
	}

	log.Debug().
		Bool("disableKeepAlive", conf.DisableKeepAlive).
		Dur("timeout", conf.Timeout).
		Int("maxIdleConnsPerHost", maxIdleConnsPerHost).
		Msg("creating http client")

	tr := &http.Transport{
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		DisableKeepAlives:   conf.DisableKeepAlive,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   conf.Timeout,
	}

	if conf.DisableFollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}
