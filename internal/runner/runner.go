package runner

import (
	"context"
	"errors"
	"fmt"
	"github.com/lameaux/bro/internal/checker"
	"github.com/lameaux/bro/internal/config"
	"github.com/lameaux/bro/internal/metrics"
	"github.com/lameaux/bro/internal/stats"
	"github.com/lameaux/bro/internal/thresholds"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type contextKey string

type Runner struct {
	httpClient      *http.Client
	scenario        *config.Scenario
	requestCounters *stats.RequestCounters
}

func New(httpClient *http.Client, scenario *config.Scenario) *Runner {
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
			Int("rps", r.scenario.Rps()).
			Int("threads", r.scenario.Threads()).
			Int("queueSize", r.scenario.QueueSize()).
			Dur("duration", r.scenario.Duration()),
	).Msg("running scenario")

	thresholds.AddScenario(r.scenario)

	queue := make(chan int, r.scenario.QueueSize())
	stop := make(chan struct{})

	startTime := time.Now()

	cancel := r.startGenerator(ctx, queue, stop)
	defer cancel()

	if err := r.runSender(ctx, queue, stop); err != nil {
		return nil, fmt.Errorf("failed sending requests: %w", err)
	}

	r.requestCounters.Duration = time.Since(startTime)
	r.requestCounters.Rps = math.Round(float64(r.requestCounters.Total.Load()) / r.requestCounters.Duration.Seconds())

	return r.requestCounters, nil
}

func (r *Runner) startGenerator(ctx context.Context, queue chan<- int, stop chan struct{}) func() {
	durationTicker := time.NewTicker(r.scenario.Duration())
	rateTicker := time.NewTicker(1 * time.Second)

	go func() {
		var num int

		generate := func() {
			for i := 0; i < r.scenario.Rps(); i++ {
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
		return
	}
	defer resp.Body.Close()

	latency := time.Since(startTime)

	checkResults, success := checker.Run(r.scenario.Checks, resp)

	err = r.logCheckResults(
		ctxWithValues,
		resp,
		latency,
		r.scenario.Checks,
		checkResults,
		success,
	)
	if err != nil {
		log.Debug().
			Int("threadId", threadId).
			Int("msgId", msgId).
			Err(err).
			Msg("failed to log check results")
		return
	}

	err = r.updateMetrics(resp, success, latency)
	if err != nil {
		log.Debug().
			Int("threadId", threadId).
			Int("msgId", msgId).
			Err(err).
			Msg("failed to update metrics")
		return
	}

	thresholds.UpdateCountersForScenario(r.scenario, checkResults)
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

func (r *Runner) logCheckResults(
	ctx context.Context,
	response *http.Response,
	latency time.Duration,
	checks []config.Check,
	results []checker.CheckResult,
	success bool,
) error {
	logEvent, err := r.makeLogEvent(ctx, response, latency)
	if err != nil {
		return fmt.Errorf("failed to make LogEvent: %w", err)
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

	return nil
}

func (r *Runner) updateMetrics(response *http.Response, success bool, latency time.Duration) error {
	if success {
		r.requestCounters.Success.Add(1)
	} else {
		r.requestCounters.Invalid.Add(1)
		r.requestCounters.Failed.Add(1)

		r.countFailedRequest("check")
	}

	labels := r.responseLabels(response, success)
	metrics.HttpResponsesTotal.With(labels).Inc()

	if err := r.requestCounters.RecordLatency(latency); err != nil {
		return fmt.Errorf("failed to record latency: %w", err)
	}

	labels = r.responseLabels(response, success)
	metrics.HttpRequestDurationSec.With(labels).Observe(latency.Seconds())

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

func (r *Runner) countFailedRequest(reason string) {
	labels := r.requestLabels()
	labels["reason"] = reason
	metrics.HttpRequestsFailedTotal.With(labels).Inc()
}
