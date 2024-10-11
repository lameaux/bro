package runner

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lameaux/bro/internal/client/checker"
	"github.com/lameaux/bro/internal/client/thresholds"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type SenderError struct {
	Reason string
	Msg    string
}

func (e *SenderError) Error() string {
	return e.Msg
}

func (r *Runner) runSender(ctx context.Context, queue <-chan int, stop <-chan struct{}) error {
	var errGrp errgroup.Group

	for t := 0; t < r.scenario.Threads(); t++ {
		threadID := t

		errGrp.Go(func() error {
			for {
				select {
				case <-stop:
					log.Debug().Int("threadID", threadID).Msg("shutting down")

					return nil
				case <-ctx.Done():
					return ctx.Err()
				case msgID, ok := <-queue:
					if !ok {
						log.Debug().Int("threadID", threadID).Msg("shutting down")

						return nil
					}

					r.processMessage(ctx, threadID, msgID)
				}
			}
		})
	}

	return errGrp.Wait() //nolint:wrapcheck
}

func (r *Runner) processMessage(ctx context.Context, threadID int, msgID int) {
	ctxWithValues := context.WithValue(ctx, contextKey("threadID"), threadID)
	ctxWithValues = context.WithValue(ctxWithValues, contextKey("msgID"), msgID)

	startTime := time.Now()

	resp, err := r.sendRequest(ctxWithValues)
	if err != nil {
		log.Debug().
			Int("threadID", threadID).
			Int("msgID", msgID).
			Err(err).
			Msg("failed to send http request")

		r.trackError(err)

		return
	}
	defer resp.Body.Close()

	latency := time.Since(startTime)

	responseChecker := checker.New(r.scenario.Checks)
	checkResults, success := responseChecker.Validate(resp)

	r.logCheckResults(
		ctxWithValues,
		resp,
		latency,
		r.scenario.Checks,
		checkResults,
		success,
	)

	r.trackResponse(resp, success, latency)

	thresholds.UpdateScenario(r.scenario, checkResults)
}

func (r *Runner) sendRequest(ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		r.scenario.HTTPRequest.Method(),
		r.scenario.HTTPRequest.URL,
		r.scenario.HTTPRequest.BodyReader(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return res, nil
}
