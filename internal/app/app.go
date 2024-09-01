package app

import (
	"context"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/lameaux/bro/internal/config"
	"github.com/lameaux/bro/internal/httpclient"
	"github.com/lameaux/bro/internal/runner"
	"github.com/lameaux/bro/internal/stats"
	"github.com/lameaux/bro/internal/thresholds"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func Run(ctx context.Context, conf *config.Config, printResults bool) bool {
	results := runScenarios(ctx, conf)

	success := processResults(results)

	if printResults {
		printResultsTable(conf, results, success)
	}

	return success
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
				Msg("failed to run scenario")
			continue
		}

		results.RequestCounters[scenario.Name] = counters

		passed, err := thresholds.ValidateScenario(scenario)
		if err != nil {
			log.Warn().
				Dict("scenario", zerolog.Dict().Str("name", scenario.Name)).
				Msg("failed to validate thresholds")
			continue
		}

		results.ThresholdsPassed[scenario.Name] = passed
	}

	results.EndTime = time.Now()
	results.TotalDuration = results.EndTime.Sub(results.StartTime)

	return results
}

func processResults(results *stats.Stats) bool {
	success := true

	for _, passed := range results.ThresholdsPassed {
		if !passed {
			success = false
			break
		}
	}

	log.Info().
		Dur("totalDuration", results.TotalDuration).
		Bool("success", success).
		Msg("result")

	return success
}

func printResultsTable(conf *config.Config, results *stats.Stats, success bool) {
	fmt.Printf("Name: %s\nPath: %s\n", conf.Name, conf.FileName)

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
			counters.Rps,
			results.ThresholdsPassed[scenario.Name],
		})

	}

	t.SetStyle(table.StyleLight)
	t.Render()

	fmt.Printf("Total duration: %v\n", results.TotalDuration)

	if success {
		fmt.Println("OK")
	} else {
		fmt.Println("Failed")
	}
}
