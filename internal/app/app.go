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
	"os"
	"text/tabwriter"
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
	fmt.Printf("Total duration: %v\n", results.EndTime.Sub(results.StartTime))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	_, _ = fmt.Fprintln(w, "Scenario\tTotal Requests\tSent\tFailed\tTimed Out\tSuccessful\tInvalid\t")

	for _, scenario := range conf.Scenarios {
		counters, ok := results.RequestCounters[scenario.Name]
		if !ok {
			log.Warn().
				Dict("scenario", zerolog.Dict().Str("name", scenario.Name)).
				Msg("missing stats")
			continue
		}

		_, _ = fmt.Fprintf(
			w,
			"%s\t%d\t%d\t%d\t%d\t%d\t%d\t\n",
			scenario.Name,
			counters.Total.Load(),
			counters.Sent.Load(),
			counters.Failed.Load(),
			counters.TimedOut.Load(),
			counters.Success.Load(),
			counters.Invalid.Load(),
		)
	}

	_ = w.Flush()
}
