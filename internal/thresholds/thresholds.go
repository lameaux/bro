package thresholds

import (
	"fmt"
	"github.com/lameaux/bro/internal/checker"
	"github.com/lameaux/bro/internal/config"
	"github.com/rs/zerolog/log"
	"sync"
)

const (
	metricChecks  = "checks"
	metricLatency = "latency"
)

type CheckCounters struct {
	mu     sync.RWMutex
	passed map[string]int64
	total  map[string]int64
}

func (cc *CheckCounters) Inc(checkType string, passed bool) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.total[checkType]++

	if passed {
		cc.passed[checkType]++
	}
}

func (cc *CheckCounters) Passed(checkType string) int64 {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	return cc.passed[checkType]
}

func (cc *CheckCounters) Total(checkType string) int64 {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	return cc.total[checkType]
}

func (cc *CheckCounters) Rate(checkType string) float64 {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	if cc.total[checkType] == 0 {
		return 0
	}

	return float64(cc.passed[checkType]) / float64(cc.total[checkType])
}

var scenarioCounters = make(map[string]*CheckCounters)

func AddScenario(scenario *config.Scenario) {
	scenarioCounters[scenario.Name] = &CheckCounters{
		passed: make(map[string]int64),
		total:  make(map[string]int64),
	}
}

func UpdateCountersForScenario(
	scenario *config.Scenario,
	results []checker.CheckResult,
) {

	checkCounters := scenarioCounters[scenario.Name]

	for i, check := range scenario.Checks {
		result := results[i]
		checkCounters.Inc(check.Type, result.Pass)
	}
}

func ValidateScenario(scenario *config.Scenario) (success bool, err error) {
	success = true

	for _, threshold := range scenario.Thresholds {
		if threshold.Metric == metricChecks {
			checkCounters, ok := scenarioCounters[scenario.Name]
			if !ok {
				return false, fmt.Errorf("missing check counters for: %v", scenario.Name)
			}

			passed := true

			rate := checkCounters.Rate(threshold.Type)

			if threshold.MinRate != nil && *threshold.MinRate > rate {
				passed = false
			}

			if threshold.MaxRate != nil && *threshold.MaxRate < rate {
				passed = false
			}

			log.Debug().
				Str("metric", threshold.Metric).
				Str("type", threshold.Type).
				Float64("rate", rate).
				Bool("passed", passed).
				Msg("threshold")

			if !passed {
				success = false
			}
		}

		if threshold.Metric == metricLatency {
			success = false
		}
	}

	return
}
