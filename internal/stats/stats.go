package stats

import (
	"github.com/HdrHistogram/hdrhistogram-go"
	"sync"
	"sync/atomic"
	"time"
)

func New() *Stats {
	return &Stats{
		StartTime:        time.Now(),
		RequestCounters:  make(map[string]*RequestCounters),
		ThresholdsPassed: make(map[string]bool),
	}
}

type Stats struct {
	StartTime     time.Time
	EndTime       time.Time
	TotalDuration time.Duration

	RequestCounters  map[string]*RequestCounters
	ThresholdsPassed map[string]bool
	Rps              float64
}

func NewRequestCounters() *RequestCounters {
	return &RequestCounters{
		latencyMillis: hdrhistogram.New(1, 1e6, 3),
	}
}

type RequestCounters struct {
	Duration time.Duration
	Rps      float64

	Total   atomic.Int64
	Sent    atomic.Int64
	Success atomic.Int64
	Failed  atomic.Int64
	Timeout atomic.Int64
	Invalid atomic.Int64

	latencyLast   time.Duration
	latencyMillis *hdrhistogram.Histogram
	mu            sync.Mutex
}

func (r *RequestCounters) RecordLatency(latency time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.latencyLast = latency
	return r.latencyMillis.RecordValue(latency.Milliseconds())
}

func (r *RequestCounters) GetLatencyAtPercentile(percentile float64) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.latencyMillis.ValueAtPercentile(percentile)
}

func (r *RequestCounters) GetLastLatency() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.latencyLast
}
