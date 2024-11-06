package app

import (
	"fmt"

	"github.com/lameaux/bro/internal/client/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (a *App) loadConfig() error {
	var conf *config.Config

	var err error

	if a.flags.URL {
		conf, err = a.makeConfigFromFlags()
	} else {
		conf, err = a.loadConfigFromFile()
	}

	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	a.conf = conf

	return nil
}

func (a *App) makeConfigFromFlags() (*config.Config, error) {
	args := a.flags.Args
	if len(args) == 0 {
		return nil, fmt.Errorf("target URL is missing. Example: %s [flags] -u <URL>", a.name) //nolint:err113
	}

	url := args[0]

	conf := &config.Config{
		Name: "Calling " + url,
		HTTPClient: config.HTTPClient{
			Timeout: a.flags.Timeout,
		},
		Scenarios: []*config.Scenario{
			{
				Name:        "flags",
				RpsRaw:      a.flags.RPS,
				DurationRaw: a.flags.Duration,
				ThreadsRaw:  a.flags.Threads,
				HTTPRequest: config.HTTPRequest{
					URL:       url,
					MethodRaw: a.flags.Method,
				},
			},
		},
	}

	log.Info().
		Dict("config", zerolog.Dict().Str("name", conf.Name)).
		Msg("config generated from flags")

	return conf, nil
}

func (a *App) loadConfigFromFile() (*config.Config, error) {
	args := a.flags.Args
	if len(args) == 0 {
		return nil, fmt.Errorf("config location is missing. Example: %s [flags] <config.yaml>", a.name) //nolint:err113
	}

	configLocation := args[0]

	conf, err := config.Load(configLocation)
	if err != nil {
		return nil, fmt.Errorf("error loading config from file: %w", err)
	}

	log.Info().
		Dict("config", zerolog.Dict().Str("name", conf.Name).Str("location", conf.FileName)).
		Msg("config loaded from location")

	return conf, nil
}
