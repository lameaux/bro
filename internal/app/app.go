package app

import (
	"context"
	"fmt"
	"github.com/Lameaux/bro/internal/config"
	"github.com/Lameaux/bro/internal/httpclient"
	"github.com/Lameaux/bro/internal/runner"
	"github.com/Lameaux/bro/internal/stats"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

func Run(ctx context.Context, conf *config.Config, showStats bool) {
	results := RunScenarios(ctx, conf)

	ProcessResults(results)

	if showStats {
		PrintStats(results)
	}
}

func RunScenarios(ctx context.Context, conf *config.Config) *stats.Stats {
	httpClient := httpclient.New(conf.HttpClient)

	log.Info().
		Str("execution", conf.Execution).
		Msg("executing scenarios... press Ctrl+C (SIGINT) or send SIGTERM to terminate.")

	results := stats.New()

	for _, scenario := range conf.Scenarios {
		r := runner.New(httpClient, scenario)
		counters, err := r.Run(ctx)
		if err != nil {
			log.Error().Err(err).
				Dict("scenario", zerolog.Dict().Str("name", scenario.Name)).
				Msgf("failed to run scenario")
		}

		results.RequestCounters[scenario.Name] = counters
	}

	results.EndTime = time.Now()

	return results
}

func ProcessResults(results *stats.Stats) {
	totalDuration := results.EndTime.Sub(results.StartTime)
	log.Info().Dur("totalDuration", totalDuration).Msg("results")
}

func PrintStats(results *stats.Stats) {
	fmt.Printf("total duration: %v\n", results.EndTime.Sub(results.StartTime))
	for scenario, counters := range results.RequestCounters {
		fmt.Printf("%s: %v\n", scenario, counters)
	}

}
