package runner

import "time"

type Listener interface {
	SetDuration(duration time.Duration)
	RecordLatency(latency time.Duration)
	IncCounter(name string, reason string)
	GetCounter(name string) int64
}

const (
	CounterTotal   = "total"
	CounterSent    = "sent"
	CounterSuccess = "success"
	CounterFailed  = "failed"
	CounterTimeout = "timeout"
	CounterInvalid = "invalid"
)

var CounterNames = []string{
	CounterTotal, CounterSent, CounterSuccess, CounterFailed, CounterTimeout, CounterInvalid,
}
