package runner

import (
	"context"
	"fmt"
	"github.com/Lameaux/bro/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net/http"
	"time"
)

func RunScenario(ctx context.Context, scenario config.Scenario) error {
	log.Info().Dict(
		"scenario",
		zerolog.Dict().
			Str("name", scenario.Name).
			Int("rate", scenario.Rate).
			Dur("interval", scenario.Interval).
			Int("vus", scenario.VUs).
			Dur("duration", scenario.Duration),
	).Msg("running scenario")

	return execute(ctx, scenario)
}

func execute(ctx context.Context, scenario config.Scenario) error {
	queue := make(chan int, scenario.VUs)

	cancel := startGenerator(ctx, scenario, queue)
	defer cancel()

	return runSender(ctx, scenario, queue)
}

func startGenerator(ctx context.Context, scenario config.Scenario, queue chan<- int) func() {
	durationTicker := time.NewTicker(scenario.Duration)
	rateTicker := time.NewTicker(scenario.Interval)

	go func() {
		var num int

		generate := func() {
			for i := 0; i < scenario.Rate; i++ {
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

func runSender(ctx context.Context, scenario config.Scenario, queue <-chan int) error {
	httpClient := NewHttpClient(scenario.Http.Client)

	var g errgroup.Group
	for vu := 0; vu < scenario.VUs; vu++ {
		id := vu
		g.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case msg, ok := <-queue:
					if !ok {
						log.Debug().Int("vuId", id).Msg("shutting down")
						return nil
					}
					if err := doHttpRequest(ctx, httpClient, scenario.Http, id, msg); err != nil {
						log.Debug().
							Int("vuId", id).
							Int("msgId", msg).
							Err(err).
							Msg("failed to send http request")
					}
				}
			}
		})
	}

	return g.Wait()
}

func doHttpRequest(ctx context.Context, httpClient *http.Client, httpConfig config.Http, id int, msg int) error {
	req, err := http.NewRequestWithContext(ctx, httpConfig.Request.Method, httpConfig.Request.Url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	logEvent := log.Debug().
		Int("vuId", id).
		Int("msgId", msg).
		Str("method", httpConfig.Request.Method).
		Str("url", httpConfig.Request.Url).
		Int("code", res.StatusCode)

	if httpConfig.Response.Code == 0 {
		logEvent.Msg("response")
		return nil
	}

	logEvent = logEvent.Int("expectedCode", httpConfig.Response.Code)

	if res.StatusCode != httpConfig.Response.Code {
		logEvent.Msg("invalid http code")
		return nil
	}

	logEvent.Msg("valid http response")

	return nil
}
