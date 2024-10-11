package runner

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/lameaux/bro/internal/client/checker"
	"github.com/lameaux/bro/internal/client/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	errMissingMsgID    = errors.New("missing msgID")
	errMissingThreadID = errors.New("missing threadID")
)

func (r *Runner) makeLogEvent(
	ctx context.Context,
	response *http.Response,
	latency time.Duration,
) (*zerolog.Event, error) {
	threadID, ok := ctx.Value(contextKey("threadID")).(int)
	if !ok {
		return nil, errMissingThreadID
	}

	msgID, ok := ctx.Value(contextKey("msgID")).(int)
	if !ok {
		return nil, errMissingMsgID
	}

	logEvent := log.Debug(). //nolint:zerologlint
					Int("threadID", threadID).
					Int("msgID", msgID).
					Str("method", r.scenario.HTTPRequest.Method()).
					Str("url", r.scenario.HTTPRequest.URL).
					Int("code", response.StatusCode).
					Int64("latency", latency.Milliseconds())

	return logEvent, nil
}

func (r *Runner) logCheckResults(
	ctx context.Context,
	response *http.Response,
	latency time.Duration,
	checks []*config.Check,
	results []checker.Result,
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
