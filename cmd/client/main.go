package main

import (
	"context"
	"github.com/lameaux/bro/internal/client/app"
	"github.com/lameaux/bro/internal/shared/signals"
	"github.com/rs/zerolog/log"
	"os"
)

const (
	appName    = "bro"
	appVersion = "v0.0.1"
)

var GitHash string

func main() {
	broApp, err := app.New(appName, appVersion, GitHash)
	if err != nil {
		log.Fatal().Err(err).Msg("app start failed")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals.Handle(false, func() {
		cancel()
	})

	os.Exit(broApp.Run(ctx))
}
