package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Lameaux/bro/internal/config"
	"github.com/Lameaux/bro/internal/runner"
	"github.com/Lameaux/bro/internal/stats"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	logo = `
 _               
 | |              
 | |__  _ __ ___  
 | '_ \| '__/ _ \ 
 | |_) | | | (_) |
 |_.__/|_|  \___/ 
`
	app     = "bro"
	version = "v0.0.1"
)

var GitHash string

func main() {
	var debug = flag.Bool("debug", false, "enable debug mode")
	var showBanner = flag.Bool("banner", true, "show banner")
	var showStats = flag.Bool("stats", true, "show stats")
	var configFile = flag.String("config", "", "path to yaml file")

	flag.Parse()

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if *showBanner {
		fmt.Print(logo)
	}

	log.Info().Str("version", version).Str("build", GitHash).Msg(app)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handleSignals(cancel)

	c := loadConfig(*configFile)

	results := runScenarios(ctx, c)

	processResults(results)

	if *showStats {
		printStats(results)
	}
}

func loadConfig(configFile string) *config.Config {
	if configFile == "" {
		log.Fatal().Msgf("--config parameter is missing. Example: %s --config myconfig.yaml", app)
	}

	c, err := config.LoadFromFile(configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("error loading config from file")
	}

	log.Debug().Str("configName", c.Name).Msgf("config loaded")

	return c
}

func handleSignals(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msgf("received signal")
		cancel()
	}()
}

func runScenarios(ctx context.Context, c *config.Config) stats.Stats {
	log.Info().Msg("executing scenarios... press Ctrl+C (SIGINT) or send SIGTERM to terminate.")

	var results stats.Stats
	results.StartTime = time.Now()

	for _, scenario := range c.Scenarios {
		err := runner.RunScenario(ctx, scenario)
		if err != nil {
			log.Fatal().Err(err).Str("scenarioName", scenario.Name).Msgf("failed to run scenario")
		}
	}

	results.EndTime = time.Now()

	return results
}

func processResults(results stats.Stats) {
	totalDuration := results.EndTime.Sub(results.StartTime)
	log.Info().Dur("totalDuration", totalDuration).Msg("results")
}

func printStats(results stats.Stats) {
	fmt.Printf("test duration: %v\n", results.EndTime.Sub(results.StartTime))

}
