package app

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (a *App) setupLog() {
	switch {
	case a.flags.Debug:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case a.flags.Silent:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if !a.flags.LogJSON {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
