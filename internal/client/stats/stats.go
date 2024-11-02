package stats

import (
	"math"
	"time"
)

func New() *Stats {
	return &Stats{
		startTime:        time.Now(),
		counters:         make(map[string]*Counters),
		thresholdsPassed: make(map[string]bool),
		duration:         make(map[string]time.Duration),
	}
}

type Stats struct {
	startTime time.Time
	endTime   time.Time

	counters         map[string]*Counters
	thresholdsPassed map[string]bool
	duration         map[string]time.Duration

	// TODO: fix parallel map writes
}

func (s *Stats) StopTimer() {
	s.endTime = time.Now()
}

func (s *Stats) TotalDuration() time.Duration {
	return s.endTime.Sub(s.startTime).Round(time.Millisecond)
}

func (s *Stats) SetCounters(scenarioName string, counters *Counters) {
	s.counters[scenarioName] = counters
}

func (s *Stats) Counters(scenarioName string) *Counters {
	return s.counters[scenarioName]
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
	total := s.Counters(scenarioName).Counter(CounterTotal)
	duration := s.Duration(scenarioName)

	return math.Round(float64(total) / duration.Seconds())
}
