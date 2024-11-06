package main

import (
	"context"
	"os"

	"github.com/lameaux/bro/internal/client/app"
	"github.com/lameaux/bro/internal/shared/signals"
	"github.com/rs/zerolog/log"
)

const (
	appName = "bro"
)

var (
	Version   string //nolint:gochecknoglobals
	BuildHash string //nolint:gochecknoglobals
	BuildDate string //nolint:gochecknoglobals
)

func main() {
	broApp, err := app.New(appName, Version, BuildHash, BuildDate)
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
