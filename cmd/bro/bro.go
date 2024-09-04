package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/lameaux/bro/internal/app"
	"github.com/lameaux/bro/internal/banner"
	"github.com/lameaux/bro/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	appName    = "bro"
	appVersion = "v0.0.1"
)

var GitHash string

func main() {
	var debug = flag.Bool("debug", false, "enable debug mode")
	var silent = flag.Bool("silent", false, "enable silent mode")
	var logJson = flag.Bool("logJson", false, "log as json")
	var skipBanner = flag.Bool("skipBanner", false, "skip banner")
	var skipResults = flag.Bool("skipResults", false, "skip results")
	var skipExitCode = flag.Bool("skipExitCode", false, "skip exit code")
	var brodAddr = flag.String("brodAddr", "", "brod address")

	flag.Parse()

	if *silent {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	} else if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if !*logJson {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if !*silent && !*skipBanner {
		fmt.Print(banner.Banner)
	}

	log.Info().Str("version", appVersion).Str("build", GitHash).Msg(appName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handleSignals(func() {
		cancel()
	})

	conf := loadConfig(flag.Args())

	if *brodAddr != "" {
		go startBrodWorker(ctx, *brodAddr)
	}

	success := app.Run(ctx, conf, !*skipResults)

	if !success && !*skipExitCode {
		os.Exit(1)
	}
}

func loadConfig(args []string) *config.Config {
	if len(args) == 0 {
		log.Fatal().Msgf("config location is missing. Example: %s [flags] <config.yaml>", appName)
	}

	configFile := args[0]

	c, err := config.LoadFromFile(configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("error loading config from file")
	}

	log.Info().
		Dict("config", zerolog.Dict().Str("name", c.Name).Str("path", c.FileName)).
		Msg("config loaded")

	return c
}

func handleSignals(shutdownFn func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msg("received signal")
		shutdownFn()
	}()
}

func startBrodWorker(ctx context.Context, addr string) {
	log.Debug().Str("brod", addr).Msg("started brod worker")

	rateTicker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-rateTicker.C:
			if err := sendCounters(ctx, addr); err != nil {
				log.Warn().Err(err).Msg("failed to send counters")
			}
		}
	}
}
