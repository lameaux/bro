package app

import (
	"context"
	"fmt"
	"github.com/lameaux/bro/internal/client/config"
	"github.com/lameaux/bro/internal/client/runner"
	"github.com/lameaux/bro/internal/client/stats"
	"github.com/lameaux/bro/internal/client/thresholds"
	"github.com/lameaux/bro/internal/shared/httpclient"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	ExitSuccess = 0
	ExitError   = 1
)

type App struct {
	appName, appVersion, appBuild string

	conf   *config.Config
	flags  *Flags
	worker *stats.BrodWorker
}

func New(appName, appVersion, appBuild string) (*App, error) {
	a := &App{
		appName:    appName,
		appVersion: appVersion,
		appBuild:   appBuild,
		flags:      ParseFlags(),
	}

	a.setupLog()
	a.printAbout()

	err := a.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return a, nil
}

func (a *App) Run(ctx context.Context) int {
	results := a.runScenarios(ctx)

	success := a.processResults(results)

	if !success && !a.flags.SkipExitCode {
		return ExitError
	}

	return ExitSuccess
}

func (a *App) runScenarios(ctx context.Context) *stats.Stats {
	results := stats.New()
	defer results.StopTimer()

	httpClient := httpclient.New(a.conf.HttpClient)

	log.Info().
		Bool("parallel", a.conf.Parallel).
		Msg("executing scenarios... press Ctrl+C (SIGINT) or send SIGTERM to terminate.")

	for _, scenario := range a.conf.Scenarios {
		counters := stats.NewRequestCounters(runner.CounterNames)

		r := runner.New(httpClient, scenario, []runner.Listener{counters})

		err := r.Run(ctx)
		if err != nil {
			log.Error().Err(err).
				Dict("scenario", zerolog.Dict().Str("name", scenario.Name)).
				Msg("failed to run scenario")
			continue
		}

		results.SetRequestCounters(scenario.Name, counters)

		passed, err := thresholds.ValidateScenario(scenario)
		if err != nil {
			log.Warn().
				Dict("scenario", zerolog.Dict().Str("name", scenario.Name)).
				Msg("failed to validate thresholds")
			continue
		}

		results.SetThresholdsPassed(scenario.Name, passed)
	}

	return results
}

func (a *App) processResults(runStats *stats.Stats) bool {
	success := runStats.AllThresholdsPassed()

	log.Info().
		Dur("totalDuration", runStats.TotalDuration()).
		Bool("success", success).
		Msg("result")

	if !a.flags.SkipResults {
		printResultsTable(a.conf, runStats, success)
	}

	return success
}
