package app

import (
	"context"
	"fmt"
	"time"

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

	conf        *config.Config
	flags       *Flags
	statsSender *stats.Sender
}

func New(appName, appVersion, appBuild string) (*App, error) {
	application := &App{
		appName:    appName,
		appVersion: appVersion,
		appBuild:   appBuild,
		flags:      ParseFlags(),
	}

	application.setupLog()
	application.printAbout()

	if err := application.loadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if application.flags.BrodAddr != "" {
		w, err := stats.NewSender(application.flags.BrodAddr, application.flags.Group)
		if err != nil {
			return nil, fmt.Errorf("failed to create stats worker: %w", err)
		}

		application.statsSender = w
	}

	return application, nil
}

func (a *App) Run(ctx context.Context) int {
	if a.statsSender != nil {
		go a.statsSender.Run(ctx)
	}

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

	httpClient := httpclient.New(a.conf.HTTPClient)

	log.Info().
		Bool("parallel", a.conf.Parallel).
		Msg("executing scenarios... press Ctrl+C (SIGINT) or send SIGTERM to terminate.")

	for _, scenario := range a.conf.Scenarios {
		localCounters := stats.NewCounters()

		listeners := []runner.StatListener{localCounters}

		if a.statsSender != nil {
			listeners = append(listeners, a.statsSender)
		}

		r := runner.New(httpClient, scenario, listeners)

		startTime := time.Now()

		err := r.Run(ctx)
		if err != nil {
			log.Error().Err(err).
				Dict("scenario", zerolog.Dict().Str("name", scenario.Name)).
				Msg("failed to run scenario")

			continue
		}

		results.SetCounters(scenario.Name, localCounters)
		results.SetDuration(scenario.Name, time.Since(startTime).Round(time.Millisecond))

		passed, err := thresholds.ValidateScenario(scenario, localCounters)
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
		formattedOutput := resultsTable(a.conf, runStats, success)

		fmt.Println(formattedOutput) //nolint:forbidigo
	}

	return success
}
