package stats

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
	"github.com/lameaux/bro/internal/client/tracking"
	"github.com/rs/zerolog/log"
)

const (
	CounterTotal   = "total"
	CounterSuccess = "success"
	CounterFailed  = "failed"
	CounterTimeout = "timeout"
	CounterInvalid = "invalid"
)

type Counters struct {
	m sync.Map

	latencyMillis *hdrhistogram.Histogram
	mu            sync.Mutex
}

func NewCounters() *Counters {
	return &Counters{
		latencyMillis: hdrhistogram.New(1, 1e6, 3), //nolint:mnd,gomnd
	}
}

func (c *Counters) recordLatency(latency time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.latencyMillis.RecordValue(latency.Milliseconds()); err != nil {
		log.Warn().Err(err).Msg("failed to record latency")
	}
}

func (c *Counters) LatencyAtPercentile(percentile float64) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.latencyMillis.ValueAtPercentile(percentile)
}

func (c *Counters) Counter(key string) int64 {
	val, ok := c.m.Load(key)
	if !ok {
		return 0
	}

	return atomic.LoadInt64(val.(*int64)) //nolint:forcetypeassert
}

func (c *Counters) incCounter(key string) {
	val, _ := c.m.LoadOrStore(key, new(int64))
	atomic.AddInt64(val.(*int64), 1) //nolint:forcetypeassert
}

func (c *Counters) TrackFailed(
	_ *tracking.RequestInfo,
	err error,
) {
	c.incCounter(CounterTotal)
	c.incCounter(CounterFailed)

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		c.incCounter(CounterTimeout)
	}
}

func (c *Counters) TrackResponse(
	_ *tracking.RequestInfo,
	success bool,
	latency time.Duration,
) {
	c.incCounter(CounterTotal)

	if success {
		c.incCounter(CounterSuccess)
	} else {
		c.incCounter(CounterInvalid)
		c.incCounter(CounterFailed)
	}

	c.recordLatency(latency)
}
