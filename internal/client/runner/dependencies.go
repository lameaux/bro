package runner

import (
	"time"

	"github.com/lameaux/bro/internal/client/tracking"
)

type StatListener interface {
	TrackFailed(
		info *tracking.RequestInfo,
		err error,
	)
	TrackResponse(
		info *tracking.RequestInfo,
		success bool,
		latency time.Duration,
	)
}
