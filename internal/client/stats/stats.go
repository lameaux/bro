package stats

import (
	"time"
)

func New() *Stats {
	return &Stats{
		startTime:        time.Now(),
		requestCounters:  make(map[string]*RequestCounters),
		thresholdsPassed: make(map[string]bool),
	}
}

type Stats struct {
	startTime time.Time
	endTime   time.Time

	requestCounters  map[string]*RequestCounters
	thresholdsPassed map[string]bool
}

func (s *Stats) StopTimer() {
	s.endTime = time.Now()
}

func (s *Stats) TotalDuration() time.Duration {
	return s.endTime.Sub(s.startTime)
}

func (s *Stats) SetRequestCounters(scenarioName string, counters *RequestCounters) {
	s.requestCounters[scenarioName] = counters
}

func (s *Stats) RequestCounters(scenarioName string) *RequestCounters {
	return s.requestCounters[scenarioName]
}

func (s *Stats) SetThresholdsPassed(scenarioName string, passed bool) {
	s.thresholdsPassed[scenarioName] = passed
}

func (s *Stats) ThresholdsPassed(scenarioName string) bool {
	return s.thresholdsPassed[scenarioName]
}

func (s *Stats) AllThresholdsPassed() bool {
	for _, passed := range s.thresholdsPassed {
		if !passed {
			return false
		}
	}

	return true
}
