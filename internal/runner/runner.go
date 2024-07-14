package runner

import (
	"context"
	"errors"
	"fmt"
	"github.com/Lameaux/bro/internal/config"
	"github.com/Lameaux/bro/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type contextKey string

type Runner struct {
	httpClient *http.Client
	scenario   config.Scenario
}

func New(httpClient *http.Client, scenario config.Scenario) *Runner {
	return &Runner{httpClient: httpClient, scenario: scenario}
}

func (r *Runner) Run(ctx context.Context) error {
	log.Info().Dict(
		"scenario",
		zerolog.Dict().
			Str("name", r.scenario.Name).
			Int("rate", r.scenario.Rate).
			Dur("interval", r.scenario.Interval).
			Int("vus", r.scenario.VUs).
			Dur("duration", r.scenario.Duration),
	).Msg("running scenario")

	queue := make(chan int, r.scenario.VUs)

	cancel := r.startGenerator(ctx, queue)
	defer cancel()

	return r.runSender(ctx, queue)
}

func (r *Runner) startGenerator(ctx context.Context, queue chan<- int) func() {
	durationTicker := time.NewTicker(r.scenario.Duration)
	rateTicker := time.NewTicker(r.scenario.Interval)

	go func() {
		var num int

		generate := func() {
			for i := 0; i < r.scenario.Rate; i++ {
				num++
				queue <- num
			}
		}
		// skip initial delay
		generate()

		for {
			select {
			case <-ctx.Done():
				close(queue)
				return
			case <-durationTicker.C:
				close(queue)
				return
			case <-rateTicker.C:
				generate()
			}
		}
	}()

	return func() {
		durationTicker.Stop()
		rateTicker.Stop()
	}
}

func (r *Runner) runSender(ctx context.Context, queue <-chan int) error {
	var g errgroup.Group
	for vu := 0; vu < r.scenario.VUs; vu++ {
		vuId := vu
		g.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case msgId, ok := <-queue:
					if !ok {
						log.Debug().Int("vuId", vuId).Msg("shutting down")
						return nil
					}

					withValues := context.WithValue(ctx, contextKey("vuId"), vuId)
					withValues = context.WithValue(withValues, contextKey("msgId"), msgId)

					resp, err := r.sendRequest(withValues)
					if err != nil {
						log.Debug().
							Int("vuId", vuId).
							Int("msgId", msgId).
							Err(err).
							Msg("failed to send http request")
						continue
					}

					if err = r.validateResponse(ctx, resp); err != nil {
						log.Debug().
							Int("vuId", vuId).
							Int("msgId", msgId).
							Err(err).
							Msg("failed to validate http response")
					}

				}
			}
		})
	}

	return g.Wait()
}

func (r *Runner) sendRequest(ctx context.Context) (*http.Response, error) {
	labels := prometheus.Labels{
		"scenario": r.scenario.Name,
		"method":   r.scenario.Request.Method,
		"url":      r.scenario.Request.Url,
	}
	metrics.HttpRequestsTotal.With(labels).Inc()

	req, err := http.NewRequestWithContext(ctx, r.scenario.Request.Method, r.scenario.Request.Url, nil)
	if err != nil {
		labels["reason"] = "invalid"
		metrics.HttpRequestsFailedTotal.With(labels).Inc()
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := r.httpClient.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
			labels["reason"] = "timeout"
		} else {
			labels["reason"] = "unknown"
		}

		metrics.HttpRequestsFailedTotal.With(labels).Inc()
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	return res, nil
}

func (r *Runner) validateResponse(ctx context.Context, response *http.Response) error {
	labels := prometheus.Labels{
		"scenario": r.scenario.Name,
		"method":   r.scenario.Request.Method,
		"url":      r.scenario.Request.Url,
	}

	vuId, ok := ctx.Value(contextKey("vuId")).(int)
	if !ok {
		return errors.New("missing vuId")
	}

	msgId, ok := ctx.Value(contextKey("msgId")).(int)
	if !ok {
		return errors.New("missing msgId")
	}

	logEvent := log.Debug().
		Int("vuId", vuId).
		Int("msgId", msgId).
		Str("method", r.scenario.Request.Method).
		Str("url", r.scenario.Request.Url).
		Int("code", response.StatusCode)

	labels["code"] = strconv.Itoa(response.StatusCode)

	expectedCode := r.scenario.Response.Code
	if expectedCode > 0 {
		logEvent = logEvent.Int("expectedCode", expectedCode)
	}

	success := expectedCode == 0 || response.StatusCode == expectedCode
	logEvent = logEvent.Bool("success", success)
	labels["success"] = strconv.FormatBool(success)

	metrics.HttpResponsesTotal.With(labels).Inc()
	logEvent.Msg("response")

	return nil
}
