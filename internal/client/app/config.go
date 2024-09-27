package app

import (
	"fmt"

	"github.com/lameaux/bro/internal/client/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (a *App) loadConfig() error {
	args := a.flags.Args

	if len(args) == 0 {
		return fmt.Errorf("config location is missing. Example: %s [flags] <config.yaml>", a.appName) //nolint:err113
	}

	configFile := args[0]

	conf, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("error loading config from file: %w", err)
	}

	log.Info().
		Dict("config", zerolog.Dict().Str("name", conf.Name).Str("path", conf.FileName)).
		Msg("config loaded")

	a.conf = conf

	return nil
}
