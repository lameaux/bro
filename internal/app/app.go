package app

import (
	"context"
	"fmt"
	"github.com/Lameaux/bro/internal/config"
	"github.com/Lameaux/bro/internal/httpclient"
	"github.com/Lameaux/bro/internal/runner"
	"github.com/Lameaux/bro/internal/stats"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	log.Info().Dur("totalDuration", totalDuration).Msg("results")
}

func printStats(conf *config.Config, results *stats.Stats) {
	totalDuration := results.EndTime.Sub(results.StartTime)
	fmt.Printf("\nTotal duration: %v\n", totalDuration)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Scenario", "Total Requests", "Sent", "Successful", "Failed", "Timeout", "Invalid", "Latency @P99", "Duration"})

	for _, scenario := range conf.Scenarios {
		counters := results.RequestCounters[scenario.Name]
		if counters == nil {
			log.Warn().
				Dict("scenario", zerolog.Dict().Str("name", scenario.Name)).
				Msg("missing stats")
			continue
		}

		t.AppendRow(table.Row{
			scenario.Name,
			counters.Total.Load(),
			counters.Sent.Load(),
			counters.Success.Load(),
			counters.Failed.Load(),
			counters.TimedOut.Load(),
			counters.Invalid.Load(),
			fmt.Sprintf("%d ms", counters.GetLatencyAtPercentile(99)),
			counters.Duration.String(),
		})

	}

	t.SetStyle(table.StyleLight)
	t.Render()
}
