package runner

import (
	"context"
	"time"
)

func startGenerator(
	ctx context.Context,
	duration time.Duration,
	startRPS int,
	targetRPS int,
	queue chan<- int,
	stop chan struct{},
) func() {
	durationTicker := time.NewTicker(duration)
	rateTicker := time.NewTicker(1 * time.Second)

	go func() {
		var num, seconds int

		generate := func() {
			seconds++

			step := int(float64(targetRPS-startRPS) * (float64(seconds) / duration.Seconds()))
			currentRps := startRPS + step

			for i := 0; i < currentRps; i++ {
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
