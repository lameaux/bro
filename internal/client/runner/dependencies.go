package runner

import "time"

type StatListener interface {
	TrackError(labels map[string]string, err error)
	TrackResponse(labels map[string]string, success bool, latency time.Duration)
}
