package runner

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lameaux/bro/internal/client/config"
	"github.com/lameaux/bro/internal/client/thresholds"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type contextKey string

type Runner struct {
	httpClient *http.Client
	scenarioID int
	scenario   *config.Scenario
	listeners  []StatListener
}

func New(
	httpClient *http.Client,
	scenarioID int,
	scenario *config.Scenario,
	listeners []StatListener,
) *Runner {
	return &Runner{
		httpClient: httpClient,
		scenarioID: scenarioID,
		scenario:   scenario,
		listeners:  listeners,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	log.Info().Dict(
		"scenario",
		zerolog.Dict().
			Str("name", r.scenario.Name).
			Int("rps", r.scenario.Rps()).
			Int("threads", r.scenario.Threads()).
			Str("duration", r.scenario.Duration().Round(time.Millisecond).String()),
	).Msg("running scenario")

	thresholds.AddScenario(r.scenario)

	if len(r.scenario.Stages) > 0 {
		return r.runRampingRate(ctx)
	}

	return r.runConstantRate(ctx)
}

func (r *Runner) runConstantRate(ctx context.Context) error {
	return r.runStage(
		ctx,
		0,
		r.scenario.Name,
		r.scenario.Threads(),
		r.scenario.Duration(),
		r.scenario.Rps(),
		r.scenario.Rps(),
	)
}

func (r *Runner) runRampingRate(ctx context.Context) error {
	previousRPS := 0

	for stageID, stage := range r.scenario.Stages {
		err := r.runStage(
			ctx,
			stageID,
			stage.Name,
			stage.Threads(),
			stage.Duration(),
			previousRPS,
			stage.Rps(),
		)
		if err != nil {
			return fmt.Errorf("failed to run stage: %w", err)
		}

		previousRPS = stage.Rps()
	}

	return nil
}

func (r *Runner) runStage(
	ctx context.Context,
	stageID int,
	name string,
	threadsCount int,
	duration time.Duration,
	startRPS int,
	targetRPS int,
) error {
	log.Info().Dict(
		"stage",
		zerolog.Dict().
			Int("stageID", stageID).
			Str("name", name).
			Int("startRPS", startRPS).
			Int("targetRPS", targetRPS).
			Int("threads", threadsCount).
			Str("duration", duration.Round(time.Millisecond).String()),
	).Msg("running stage")

	queue := make(chan int, threadsCount)
	stop := make(chan struct{})

	cancel := startGenerator(
		ctx,
		duration,
		startRPS,
		targetRPS,
		queue,
		stop,
	)
	defer cancel()

	if err := r.runSender(ctx, threadsCount, queue, stop); err != nil {
		return fmt.Errorf("failed sending requests: %w", err)
	}

	return nil
}
