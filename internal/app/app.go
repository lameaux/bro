package app

import (
	"context"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/lameaux/bro/internal/config"
	"github.com/lameaux/bro/internal/httpclient"
	"github.com/lameaux/bro/internal/runner"
	"github.com/lameaux/bro/internal/stats"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math"
	"os"
	"time"
)

func Run(ctx context.Context, conf *config.Config, showStats bool) {
	results := runScenarios(ctx, conf)

	processResults(results)

	if showStats {
		printStats(conf, results)
	}
}

func runScenarios(ctx context.Context, conf *config.Config) *stats.Stats {
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

func processResults(results *stats.Stats) {
	totalDuration := results.EndTime.Sub(results.StartTime)
	log.Info().Dur("totalDuration", totalDuration).Bool("ok", true).Msg("results")
}

func printStats(conf *config.Config, results *stats.Stats) {
	totalDuration := results.EndTime.Sub(results.StartTime)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Scenario", "Total", "Sent", "Success", "Failed", "Timeout", "Invalid", "Latency @P99", "Duration", "RPS", "Passed"})

	for _, scenario := range conf.Scenarios {
		counters := results.RequestCounters[scenario.Name]
		if counters == nil {
			log.Warn().
				Dict("scenario", zerolog.Dict().Str("name", scenario.Name)).
				Msg("missing stats")
			continue
		}

		rps := math.Round(float64(counters.Total.Load()) / counters.Duration.Seconds())

		t.AppendRow(table.Row{
			scenario.Name,
			counters.Total.Load(),
			counters.Sent.Load(),
			counters.Success.Load(),
			counters.Failed.Load(),
			counters.Timeout.Load(),
			counters.Invalid.Load(),
			fmt.Sprintf("%d ms", counters.GetLatencyAtPercentile(99)),
			counters.Duration,
			rps,
			"OK",
		})

	}

	t.SetStyle(table.StyleLight)
	t.Render()

	fmt.Printf("Total duration: %v\n", totalDuration)
	fmt.Println("OK")
}
