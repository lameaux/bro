package runner

import (
	"context"
	"errors"
	"fmt"
	"github.com/lameaux/bro/internal/checker"
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
			Int("rps", r.scenario.RPS).
			Int("vus", r.scenario.VUs).
			Dur("duration", r.scenario.Duration),
	).Msg("running scenario")

	queue := make(chan int, max(r.scenario.VUs, r.scenario.Buffer))
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
	rateTicker := time.NewTicker(1 * time.Second)

	go func() {
		var num int

		generate := func() {
			for i := 0; i < r.scenario.RPS; i++ {
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

					r.processMessage(ctx, vuId, msgId)
				}
			}
		})
	}

	return g.Wait()
}

func (r *Runner) processMessage(ctx context.Context, vuId int, msgId int) {
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
		return
	}
	defer resp.Body.Close()

	latency := time.Since(startTime)

	if err = r.runChecks(ctxWithValues, resp, latency); err != nil {
		log.Debug().
			Int("vuId", vuId).
			Int("msgId", msgId).
			Err(err).
			Msg("failed to run checks")
	}

	// validate threshold
	// TODO
}

func (r *Runner) sendRequest(ctx context.Context) (*http.Response, error) {
	r.requestCounters.Sent.Add(1)
	metrics.HttpRequestsTotal.With(r.requestLabels()).Inc()

	req, err := http.NewRequestWithContext(
		ctx,
		r.scenario.HttpRequest.Method(),
		r.scenario.HttpRequest.Url,
		r.scenario.HttpRequest.BodyReader(),
	)
	if err != nil {
		r.requestCounters.Failed.Add(1)
		r.countFailedRequest("format")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := r.httpClient.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
			r.requestCounters.Timeout.Add(1)
			r.countFailedRequest("timeout")
		} else {
			r.requestCounters.Failed.Add(1)
			r.countFailedRequest("unknown")
		}

		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return res, nil
}

func (r *Runner) runChecks(ctx context.Context, response *http.Response, latency time.Duration) error {
	logEvent, err := r.makeLogEvent(ctx, response, latency)
	if err != nil {
		return fmt.Errorf("failed to make LogEvent: %w", err)
	}

	success := true
	checkResults := zerolog.Arr()
	for _, check := range r.scenario.Checks {
		value, ok, err := checker.Run(check, response)
		checkResults = checkResults.Dict(
			zerolog.Dict().
				Str("type", check.Type).
				Str("name", check.Name).
				Str("value", value).
				Bool("ok", ok).
				Err(err),
		)
		if !ok {
			success = false
		}
	}

	logEvent.Array("checks", checkResults)

	if success {
		r.requestCounters.Success.Add(1)
	} else {
		r.requestCounters.Invalid.Add(1)
		r.requestCounters.Failed.Add(1)

		r.countFailedRequest("code")
	}

	labels := r.responseLabels(response, success)
	metrics.HttpResponsesTotal.With(labels).Inc()

	if err = r.requestCounters.RecordLatency(latency); err != nil {
		return fmt.Errorf("failed to record latency: %w", err)
	}

	labels = r.responseLabels(response, success)
	metrics.HttpRequestDurationSec.With(labels).Observe(latency.Seconds())

	logEvent.Bool("success", success).Msg("response")

	return nil
}

func (r *Runner) requestLabels() prometheus.Labels {
	return prometheus.Labels{
		"scenario": r.scenario.Name,
		"method":   r.scenario.HttpRequest.Method(),
		"url":      r.scenario.HttpRequest.Url,
	}
}

func (r *Runner) responseLabels(response *http.Response, success bool) prometheus.Labels {
	labels := r.requestLabels()
	labels["code"] = strconv.Itoa(response.StatusCode)
	labels["success"] = strconv.FormatBool(success)

	return labels
}

func (r *Runner) makeLogEvent(
	ctx context.Context,
	response *http.Response,
	latency time.Duration,
) (*zerolog.Event, error) {
	vuId, ok := ctx.Value(contextKey("vuId")).(int)
	if !ok {
		return nil, errors.New("missing vuId")
	}

	msgId, ok := ctx.Value(contextKey("msgId")).(int)
	if !ok {
		return nil, errors.New("missing msgId")
	}

	logEvent := log.Debug().
		Int("vuId", vuId).
		Int("msgId", msgId).
		Str("method", r.scenario.HttpRequest.Method()).
		Str("url", r.scenario.HttpRequest.Url).
		Int("code", response.StatusCode).
		Int64("latency", latency.Milliseconds())

	return logEvent, nil
}

func (r *Runner) countFailedRequest(reason string) {
	labels := r.requestLabels()
	labels["reason"] = reason
	metrics.HttpRequestsFailedTotal.With(labels).Inc()
}
