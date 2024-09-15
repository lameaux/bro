package runner

import (
	"context"
	"errors"
	"github.com/lameaux/bro/internal/client/checker"
	"github.com/lameaux/bro/internal/client/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

func (r *Runner) makeLogEvent(
	ctx context.Context,
	response *http.Response,
	latency time.Duration,
) (*zerolog.Event, error) {
	threadId, ok := ctx.Value(contextKey("threadId")).(int)
	if !ok {
		return nil, errors.New("missing threadId")
	}

	msgId, ok := ctx.Value(contextKey("msgId")).(int)
	if !ok {
		return nil, errors.New("missing msgId")
	}

	logEvent := log.Debug().
		Int("threadId", threadId).
		Int("msgId", msgId).
		Str("method", r.scenario.HttpRequest.Method()).
		Str("url", r.scenario.HttpRequest.Url).
		Int("code", response.StatusCode).
		Int64("latency", latency.Milliseconds())

	return logEvent, nil
}

func (r *Runner) logCheckResults(
	ctx context.Context,
	response *http.Response,
	latency time.Duration,
	checks []config.Check,
	results []checker.CheckResult,
	success bool,
) {
	logEvent, err := r.makeLogEvent(ctx, response, latency)
	if err != nil {
		log.Warn().Err(err).Msg("failed to log check results")
		return
	}

	checkResults := zerolog.Arr()
	for i, check := range checks {
		result := results[i]

		checkResults = checkResults.Dict(
			zerolog.Dict().
				Str("type", check.Type).
				Str("name", check.Name).
				Str("value", result.Actual).
				Bool("pass", result.Pass).
				Err(result.Error),
		)
	}

	logEvent.
		Array("checks", checkResults).
		Bool("success", success).
		Msg("response")
}
