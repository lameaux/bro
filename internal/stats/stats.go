package stats

import (
	"sync/atomic"
	"time"
)

func New() *Stats {
	return &Stats{
		StartTime:       time.Now(),
		RequestCounters: make(map[string]*RequestCounters),
	}
}

type Stats struct {
	StartTime       time.Time
	EndTime         time.Time
	RequestCounters map[string]*RequestCounters
}

type RequestCounters struct {
	Total    atomic.Int64
	Sent     atomic.Int64
	Failed   atomic.Int64
	TimedOut atomic.Int64
	Success  atomic.Int64
	Invalid  atomic.Int64
}
