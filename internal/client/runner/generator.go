package runner

import (
	"context"
	"time"
)

func (r *Runner) startGenerator(ctx context.Context, queue chan<- int, stop chan struct{}) func() {
	durationTicker := time.NewTicker(r.scenario.Duration())
	rateTicker := time.NewTicker(1 * time.Second)

	go func() {
		var num int

		generate := func() {
			for i := 0; i < r.scenario.Rps(); i++ {
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
