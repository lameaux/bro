package runner

import (
	"context"
	"fmt"
	"github.com/lameaux/bro/internal/client/checker"
	"github.com/lameaux/bro/internal/client/thresholds"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net/http"
	"time"
)

type SenderError struct {
	Reason string
	Msg    string
}

func (e *SenderError) Error() string {
	return e.Msg
}

func (r *Runner) runSender(ctx context.Context, queue <-chan int, stop <-chan struct{}) error {
	var g errgroup.Group
	for t := 0; t < r.scenario.Threads(); t++ {
		threadId := t
		g.Go(func() error {
			for {
				select {
				case <-stop:
					log.Debug().Int("threadId", threadId).Msg("shutting down")
					return nil
				case <-ctx.Done():
					return ctx.Err()
				case msgId, ok := <-queue:
					if !ok {
						log.Debug().Int("threadId", threadId).Msg("shutting down")
						return nil
					}

					r.processMessage(ctx, threadId, msgId)
				}
			}
		})
	}

	return g.Wait()
}

func (r *Runner) processMessage(ctx context.Context, threadId int, msgId int) {
	ctxWithValues := context.WithValue(ctx, contextKey("threadId"), threadId)
	ctxWithValues = context.WithValue(ctxWithValues, contextKey("msgId"), msgId)

	startTime := time.Now()

	resp, err := r.sendRequest(ctxWithValues)
	if err != nil {
		log.Debug().
			Int("threadId", threadId).
			Int("msgId", msgId).
			Err(err).
			Msg("failed to send http request")

		r.trackError(err)
		return
	}
	defer resp.Body.Close()

	latency := time.Since(startTime)

	checkResults, success := checker.Run(r.scenario.Checks, resp)

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
		r.scenario.HttpRequest.Method(),
		r.scenario.HttpRequest.Url,
		r.scenario.HttpRequest.BodyReader(),
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
