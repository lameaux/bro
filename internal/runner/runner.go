package runner

import (
	"context"
	"fmt"
	"github.com/Lameaux/bro/internal/config"
	"golang.org/x/sync/errgroup"
	"net/http"
	"time"
)

func RunScenario(ctx context.Context, scenario config.Scenario) error {
	fmt.Printf("Scenario: %s\n", scenario.Name)
	fmt.Printf("Generating %d requests per %v, running %d VUs for %v\n",
		scenario.Rate,
		scenario.Interval,
		scenario.VUs,
		scenario.Duration,
	)
	return execute(ctx, scenario)
}

func execute(ctx context.Context, scenario config.Scenario) error {
	durationTicker := time.NewTicker(scenario.Duration)
	defer durationTicker.Stop()

	rateTicker := time.NewTicker(scenario.Interval)
	defer rateTicker.Stop()

	httpClient := NewHttpClient(scenario.Request.Http)

	queue := make(chan int, scenario.VUs)

	// producer
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

	// consumers
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
						fmt.Printf("VU %d: shutting down\n", id)
						return nil
					}
					if err := doHttpRequest(ctx, httpClient, scenario.Request.Http, id, msg); err != nil {
						fmt.Printf("VU %d, msg %d: failed to send http request: %v\n", id, msg, err)
					}
				}
			}
		})
	}

	return g.Wait()
}

func doHttpRequest(ctx context.Context, httpClient *http.Client, request config.HttpRequest, id int, msg int) error {
	req, err := http.NewRequestWithContext(ctx, request.Method, request.Url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	fmt.Printf("VU %d, msg %d: %d\n", id, msg, res.StatusCode)

	return nil
}
