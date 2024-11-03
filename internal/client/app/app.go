package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
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
	exitSuccess = 0
	exitError   = 1

	outputFilePermissions = 0o640
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
		return exitError
	}

	return exitSuccess
}

func (a *App) runScenarios(ctx context.Context) *stats.Stats {
	results := stats.New()
	defer results.StopTimer()

	httpClient := httpclient.New(a.conf.HTTPClient)

	log.Info().
		Bool("parallel", a.conf.Parallel).
		Msg("executing scenarios... press Ctrl+C (SIGINT) or send SIGTERM to terminate.")

	if a.conf.Parallel {
		a.runParallel(ctx, httpClient, results)
	} else {
		a.runSerial(ctx, httpClient, results)
	}

	return results
}

func (a *App) runParallel(
	ctx context.Context,
	httpClient *http.Client,
	results *stats.Stats,
) {
	var wg sync.WaitGroup

	wg.Add(len(a.conf.Scenarios))

	for scenarioID, scenario := range a.conf.Scenarios {
		go func(scenarioID int, scenario *config.Scenario) {
			defer wg.Done()
			a.runScenario(ctx, httpClient, scenarioID, scenario, results)
		}(scenarioID, scenario)
	}

	wg.Wait()
}

func (a *App) runSerial(
	ctx context.Context,
	httpClient *http.Client,
	results *stats.Stats,
) {
	for scenarioID, scenario := range a.conf.Scenarios {
		a.runScenario(ctx, httpClient, scenarioID, scenario, results)
	}
}

func (a *App) runScenario(
	ctx context.Context,
	httpClient *http.Client,
	scenarioID int,
	scenario *config.Scenario,
	results *stats.Stats,
) {
	localCounters := stats.NewCounters()

	listeners := []runner.StatListener{localCounters}

	if a.statsSender != nil {
		listeners = append(listeners, a.statsSender)
	}

	r := runner.New(httpClient, scenarioID, scenario, listeners)

	startTime := time.Now()

	err := r.Run(ctx)
	if err != nil {
		log.Error().Err(err).
			Dict("scenario", zerolog.Dict().Str("name", scenario.Name)).
			Msg("failed to run scenario")

		return
	}

	results.SetCounters(scenario.Name, localCounters)
	results.SetDuration(scenario.Name, time.Since(startTime).Round(time.Millisecond))

	passed, err := thresholds.ValidateScenario(scenario, localCounters)
	if err != nil {
		log.Warn().
			Dict("scenario", zerolog.Dict().Str("name", scenario.Name)).
			Msg("failed to validate thresholds")

		return
	}

	results.SetThresholdsPassed(scenario.Name, passed)
}

func (a *App) processResults(runStats *stats.Stats) bool {
	success := runStats.AllThresholdsPassed()

	log.Info().
		Str("totalDuration", runStats.TotalDuration().Round(time.Millisecond).String()).
		Bool("success", success).
		Msg("result")

	var formattedOutput string

	switch a.flags.Format {
	case "txt":
		formattedOutput = generateTXT(a.conf, runStats, success)
	case "csv":
		formattedOutput = generateCSV(a.conf, runStats)
	default:
		log.Error().Str("format", a.flags.Format).Msg("invalid format")
	}

	if a.flags.Output == "stdout" {
		fmt.Println(formattedOutput) //nolint:forbidigo

		return success
	}

	err := os.WriteFile(a.flags.Output, []byte(formattedOutput), outputFilePermissions)
	if err != nil {
		log.Error().Str("output", a.flags.Output).Err(err).Msg("failed to print output")
	}

	return success
}
