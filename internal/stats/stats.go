package stats

import (
	"sync/atomic"
	"time"
)

type Stats struct {
	StartTime    time.Time
	EndTime      time.Time
	ScenarioStat []ScenarioStat
}

type ScenarioStat struct {
	TotalRequests    atomic.Int64
	FailedRequests   atomic.Int64
	TimedOutRequests atomic.Int64
}
