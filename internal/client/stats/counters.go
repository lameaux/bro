package stats

import (
	"github.com/HdrHistogram/hdrhistogram-go"
	"github.com/lameaux/bro/internal/client/runner"
	"github.com/rs/zerolog/log"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

type RequestCounters struct {
	Duration time.Duration

	counters map[string]*atomic.Int64

	latencyMillis *hdrhistogram.Histogram
	mu            sync.Mutex
}

func NewRequestCounters(counterNames []string) *RequestCounters {
	m := make(map[string]*atomic.Int64, len(counterNames))

	for _, name := range counterNames {
		var c atomic.Int64
		m[name] = &c
	}

	rc := &RequestCounters{
		counters:      m,
		latencyMillis: hdrhistogram.New(1, 1e6, 3),
	}

	return rc
}

func (r *RequestCounters) RecordLatency(latency time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.latencyMillis.RecordValue(latency.Milliseconds()); err != nil {
		log.Warn().Err(err).Msg("failed to record latency")
	}
}

func (r *RequestCounters) GetLatencyAtPercentile(percentile float64) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.latencyMillis.ValueAtPercentile(percentile)
}

func (r *RequestCounters) SetDuration(duration time.Duration) {
	r.Duration = duration
}

func (r *RequestCounters) Rps() float64 {
	return math.Round(float64(r.GetCounter(runner.CounterTotal)) / r.Duration.Seconds())
}

func (r *RequestCounters) IncCounter(name string, _ string) {
	r.counters[name].Add(1)
}

func (r *RequestCounters) GetCounter(name string) int64 {
	return r.counters[name].Load()
}
