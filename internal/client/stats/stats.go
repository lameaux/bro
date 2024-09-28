package stats

import (
	"math"
	"time"
)

func New() *Stats {
	return &Stats{
		startTime:        time.Now(),
		requestCounters:  make(map[string]*RequestCounters),
		thresholdsPassed: make(map[string]bool),
		duration:         make(map[string]time.Duration),
	}
}

type Stats struct {
	startTime time.Time
	endTime   time.Time

	requestCounters  map[string]*RequestCounters
	thresholdsPassed map[string]bool
	duration         map[string]time.Duration
}

func (s *Stats) StopTimer() {
	s.endTime = time.Now()
}

func (s *Stats) TotalDuration() time.Duration {
	return s.endTime.Sub(s.startTime).Round(time.Millisecond)
}

func (s *Stats) SetRequestCounters(scenarioName string, counters *RequestCounters) {
	s.requestCounters[scenarioName] = counters
}

func (s *Stats) RequestCounters(scenarioName string) *RequestCounters {
	return s.requestCounters[scenarioName]
}

func (s *Stats) SetDuration(scenarioName string, d time.Duration) {
	s.duration[scenarioName] = d
}

func (s *Stats) Duration(scenarioName string) time.Duration {
	return s.duration[scenarioName]
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

func (s *Stats) Rps(scenarioName string) float64 {
	total := s.RequestCounters(scenarioName).Counter(CounterTotal)
	duration := s.Duration(scenarioName)

	return math.Round(float64(total) / duration.Seconds())
}
