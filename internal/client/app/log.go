package app

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func (a *App) setupLog() {
	if a.flags.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else if a.flags.Silent {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if !a.flags.LogJson {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
