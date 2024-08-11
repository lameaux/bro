package runner

import (
	"context"
	"errors"
	"fmt"
	"github.com/lameaux/bro/internal/config"
	"github.com/lameaux/bro/internal/metrics"
	"github.com/lameaux/bro/internal/stats"
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
	httpClient      *http.Client
	scenario        config.Scenario
	requestCounters *stats.RequestCounters
}

func New(httpClient *http.Client, scenario config.Scenario) *Runner {
	return &Runner{
		httpClient:      httpClient,
		scenario:        scenario,
		requestCounters: stats.NewRequestCounters(),
	}
}

func (r *Runner) Run(ctx context.Context) (*stats.RequestCounters, error) {
	log.Info().Dict(
		"scenario",
		zerolog.Dict().
			Str("name", r.scenario.Name).
			Int("rate", r.scenario.Rate).
			Dur("interval", r.scenario.Interval).
			Int("vus", r.scenario.VUs).
			Dur("duration", r.scenario.Duration),
	).Msg("running scenario")

	queue := make(chan int, maxInt(r.scenario.VUs, r.scenario.Buffer))
	stop := make(chan struct{})

	startTime := time.Now()

	cancel := r.startGenerator(ctx, queue, stop)
	defer cancel()

	if err := r.runSender(ctx, queue, stop); err != nil {
		return nil, fmt.Errorf("failed sending requests: %w", err)
	}

	r.requestCounters.Duration = time.Since(startTime)

	return r.requestCounters, nil
}

func (r *Runner) startGenerator(ctx context.Context, queue chan<- int, stop chan struct{}) func() {
	durationTicker := time.NewTicker(r.scenario.Duration)
	rateTicker := time.NewTicker(r.scenario.Interval)

	go func() {
		var num int

		generate := func() {
			for i := 0; i < r.scenario.Rate; i++ {
				num++
				r.requestCounters.Total.Add(1)
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
				close(stop)
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

func (r *Runner) runSender(ctx context.Context, queue <-chan int, stop <-chan struct{}) error {
	var g errgroup.Group
	for vu := 0; vu < r.scenario.VUs; vu++ {
		vuId := vu
		g.Go(func() error {
			for {
				select {
				case <-stop:
					log.Debug().Int("vuId", vuId).Msg("shutting down")
					return nil
				case <-ctx.Done():
					return ctx.Err()
				case msgId, ok := <-queue:
					if !ok {
						log.Debug().Int("vuId", vuId).Msg("shutting down")
						return nil
					}

					ctxWithValues := context.WithValue(ctx, contextKey("vuId"), vuId)
					ctxWithValues = context.WithValue(ctxWithValues, contextKey("msgId"), msgId)

					startTime := time.Now()
					resp, err := r.sendRequest(ctxWithValues)
					if err != nil {
						log.Debug().
							Int("vuId", vuId).
							Int("msgId", msgId).
							Err(err).
							Msg("failed to send http request")
						continue
					}

					latency := time.Since(startTime)

					if err = r.validateResponse(ctxWithValues, resp, latency); err != nil {
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
		"method":   r.scenario.HttpRequest.Method,
		"url":      r.scenario.HttpRequest.Url,
	}

	metrics.HttpRequestsTotal.With(labels).Inc()
	r.requestCounters.Sent.Add(1)

	req, err := http.NewRequestWithContext(ctx, r.scenario.HttpRequest.Method, r.scenario.HttpRequest.Url, nil)
	if err != nil {
		labels["reason"] = "invalid"
		metrics.HttpRequestsFailedTotal.With(labels).Inc()
		r.requestCounters.Failed.Add(1)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := r.httpClient.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
			labels["reason"] = "timeout"
			r.requestCounters.TimedOut.Add(1)
		} else {
			labels["reason"] = "unknown"
			r.requestCounters.Failed.Add(1)
		}

		metrics.HttpRequestsFailedTotal.With(labels).Inc()
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	return res, nil
}

func (r *Runner) validateResponse(ctx context.Context, response *http.Response, latency time.Duration) error {
	labels := prometheus.Labels{
		"scenario": r.scenario.Name,
		"method":   r.scenario.HttpRequest.Method,
		"url":      r.scenario.HttpRequest.Url,
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
		Str("method", r.scenario.HttpRequest.Method).
		Str("url", r.scenario.HttpRequest.Url).
		Int("code", response.StatusCode).
		Int64("latency", latency.Milliseconds())

	labels["code"] = strconv.Itoa(response.StatusCode)

	expectedCode := r.scenario.HttpResponse.Code
	if expectedCode > 0 {
		logEvent = logEvent.Int("expectedCode", expectedCode)
	}

	success := expectedCode == 0 || response.StatusCode == expectedCode
	logEvent = logEvent.Bool("success", success)
	labels["success"] = strconv.FormatBool(success)

	metrics.HttpResponsesTotal.With(labels).Inc()

	if success {
		r.requestCounters.Success.Add(1)
	} else {
		r.requestCounters.Invalid.Add(1)
	}

	metrics.HttpRequestDurationSec.With(labels).Observe(latency.Seconds())

	if err := r.requestCounters.RecordLatency(latency); err != nil {
		return fmt.Errorf("failed to record latency: %w", err)
	}

	logEvent.Msg("response")

	return nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}
