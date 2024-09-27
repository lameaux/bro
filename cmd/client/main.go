package main

import (
	"context"
	"os"

	"github.com/lameaux/bro/internal/client/app"
	"github.com/lameaux/bro/internal/shared/signals"
	"github.com/rs/zerolog/log"
)

const (
	appName    = "bro"
	appVersion = "v0.0.1"
)

var GitHash string //nolint:gochecknoglobals

func main() {
	broApp, err := app.New(appName, appVersion, GitHash)
	if err != nil {
		log.Fatal().Err(err).Msg("app start failed")

		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	signals.Handle(false, func() {
		cancel()
	})

	os.Exit(broApp.Run(ctx))
}
