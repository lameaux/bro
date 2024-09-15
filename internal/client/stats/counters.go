package stats

import (
	"errors"
	"github.com/HdrHistogram/hdrhistogram-go"
	"github.com/rs/zerolog/log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	CounterTotal   = "total"
	CounterSuccess = "success"
	CounterFailed  = "failed"
	CounterTimeout = "timeout"
	CounterInvalid = "invalid"
)

type RequestCounters struct {
	counters sync.Map

	latencyMillis *hdrhistogram.Histogram
	mu            sync.Mutex
}

func NewRequestCounters() *RequestCounters {
	rc := &RequestCounters{
		latencyMillis: hdrhistogram.New(1, 1e6, 3),
	}

	return rc
}

func (r *RequestCounters) recordLatency(latency time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.latencyMillis.RecordValue(latency.Milliseconds()); err != nil {
		log.Warn().Err(err).Msg("failed to record latency")
	}
}

func (r *RequestCounters) LatencyAtPercentile(percentile float64) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.latencyMillis.ValueAtPercentile(percentile)
}

//func (r *RequestCounters) Rps() float64 {
//	return math.Round(float64(r.Counter(runner.CounterTotal)) / r.duration.Seconds())
//}

func (r *RequestCounters) Counter(key string) int64 {
	val, ok := r.counters.Load(key)
	if !ok {
		return 0
	}
	return atomic.LoadInt64(val.(*int64))
}

func (r *RequestCounters) incCounter(key string) {
	val, _ := r.counters.LoadOrStore(key, new(int64))
	atomic.AddInt64(val.(*int64), 1)
}

func (r *RequestCounters) TrackError(_labels map[string]string, err error) {
	r.incCounter(CounterTotal)
	r.incCounter(CounterFailed)

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		r.incCounter(CounterTimeout)
	}
}

func (r *RequestCounters) TrackResponse(_labels map[string]string, success bool, latency time.Duration) {
	r.incCounter(CounterTotal)

	if success {
		r.incCounter(CounterSuccess)
	} else {
		r.incCounter(CounterInvalid)
		r.incCounter(CounterFailed)
	}

	r.recordLatency(latency)
}
