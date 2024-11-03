package stats

import (
	"math"
	"sync"
	"time"
)

func New() *Stats {
	return &Stats{
		startTime: time.Now(),
	}
}

type Stats struct {
	startTime time.Time
	endTime   time.Time

	counters         sync.Map // *Counters
	passedThresholds sync.Map // bool
	durations        sync.Map // time.Duration
}

func (s *Stats) StopTimer() {
	s.endTime = time.Now()
}

func (s *Stats) TotalDuration() time.Duration {
	return s.endTime.Sub(s.startTime).Round(time.Millisecond)
}

func (s *Stats) SetCounters(scenarioName string, counters *Counters) {
	s.counters.Store(scenarioName, counters)
}

func (s *Stats) Counters(scenarioName string) *Counters {
	value, ok := s.counters.Load(scenarioName)
	if !ok {
		return nil
	}

	c, _ := value.(*Counters)

	return c
}

func (s *Stats) SetDuration(scenarioName string, d time.Duration) {
	s.durations.Store(scenarioName, d)
}

func (s *Stats) Duration(scenarioName string) time.Duration {
	value, ok := s.durations.Load(scenarioName)
	if !ok {
		return 0
	}

	d, _ := value.(time.Duration)

	return d
}

func (s *Stats) SetThresholdsPassed(scenarioName string, passed bool) {
	s.passedThresholds.Store(scenarioName, passed)
}

func (s *Stats) ThresholdsPassed(scenarioName string) bool {
	value, ok := s.passedThresholds.Load(scenarioName)
	if !ok {
		return false
	}

	passed, _ := value.(bool)

	return passed
}

func (s *Stats) AllThresholdsPassed() bool {
	passed := true

	s.passedThresholds.Range(func(_, value any) bool {
		passed, _ = value.(bool)

		return passed
	})

	return passed
}

func (s *Stats) Rps(scenarioName string) float64 {
	total := s.Counters(scenarioName).Counter(CounterTotal)
	duration := s.Duration(scenarioName)

	return math.Round(float64(total) / duration.Seconds())
}
