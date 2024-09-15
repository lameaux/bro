package runner

import (
	"context"
	"fmt"
	"github.com/lameaux/bro/internal/client/config"
	"github.com/lameaux/bro/internal/client/thresholds"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
)

type contextKey string

type Runner struct {
	httpClient *http.Client
	scenario   *config.Scenario
	listeners  []StatListener
}

func New(httpClient *http.Client, scenario *config.Scenario, listeners []StatListener) *Runner {
	return &Runner{
		httpClient: httpClient,
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
			Int("queueSize", r.scenario.QueueSize()).
			Dur("duration", r.scenario.Duration()),
	).Msg("running scenario")

	thresholds.AddScenario(r.scenario)

	queue := make(chan int, r.scenario.QueueSize())
	stop := make(chan struct{})

	cancel := r.startGenerator(ctx, queue, stop)
	defer cancel()

	if err := r.runSender(ctx, queue, stop); err != nil {
		return fmt.Errorf("failed sending requests: %w", err)
	}

	return nil
}
