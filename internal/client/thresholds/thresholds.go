package thresholds

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/lameaux/bro/internal/client/checker"
	"github.com/lameaux/bro/internal/client/config"
	"github.com/lameaux/bro/internal/client/stats"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	metricChecks  = "checks"
	metricLatency = "latency"
)

var errMissingCheckCounters = errors.New("missing check counters")

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

// FIXME: refactor into struct.
var scenarioCounters = make(map[string]*CheckCounters) //nolint:gochecknoglobals

func AddScenario(scenario *config.Scenario) {
	scenarioCounters[scenario.Name] = &CheckCounters{
		passed: make(map[string]int64),
		total:  make(map[string]int64),
	}
}

func UpdateScenario(
	scenario *config.Scenario,
	results []checker.CheckResult,
) {
	checkCounters := scenarioCounters[scenario.Name]

	for i, check := range scenario.Checks {
		result := results[i]
		checkCounters.Inc(check.Type, result.Pass)
	}
}

func ValidateScenario(
	scenario *config.Scenario,
	counters *stats.Counters,
) (bool, error) {
	success := true

	for _, threshold := range scenario.Thresholds {
		if threshold.Metric == metricChecks {
			passed, err := validateMetricCheck(scenario, threshold)
			if err != nil {
				return false, fmt.Errorf("failed to validate metric check: %w", err)
			}

			if !passed {
				success = false
			}
		}

		if threshold.Metric == metricLatency {
			passed, err := validateLatencyCheck(scenario, threshold, counters)
			if err != nil {
				return false, fmt.Errorf("failed to validate latency check: %w", err)
			}

			if !passed {
				success = false
			}
		}
	}

	return success, nil
}

func validateMetricCheck(
	scenario *config.Scenario,
	threshold *config.Threshold,
) (bool, error) {
	checkCounters, ok := scenarioCounters[scenario.Name]
	if !ok {
		return false, errMissingCheckCounters
	}

	passed := true

	rate := checkCounters.Rate(threshold.Type)

	if threshold.MinRate != nil && *threshold.MinRate > rate {
		passed = false
	}

	if threshold.MaxRate != nil && *threshold.MaxRate < rate {
		passed = false
	}

	count := checkCounters.Passed(threshold.Type)

	if threshold.MinCount != nil && *threshold.MinCount > count {
		passed = false
	}

	if threshold.MaxCount != nil && *threshold.MaxCount < count {
		passed = false
	}

	logThresholdValidation(scenario, threshold, passed, rate, count, 0)

	return passed, nil
}

func validateLatencyCheck(
	scenario *config.Scenario,
	threshold *config.Threshold,
	counters *stats.Counters,
) (bool, error) {
	percentile, err := strconv.ParseFloat(threshold.Type, 64)
	if err != nil {
		return false, fmt.Errorf("invalid percentile: %w", err)
	}

	passed := true

	value := float64(counters.LatencyAtPercentile(percentile))

	if threshold.MinValue != nil && *threshold.MinValue > value {
		passed = false
	}

	if threshold.MaxValue != nil && *threshold.MaxValue < value {
		passed = false
	}

	logThresholdValidation(scenario, threshold, passed, 0, 0, value)

	return passed, nil
}

func logThresholdValidation(
	scenario *config.Scenario,
	threshold *config.Threshold,
	passed bool,
	rate float64,
	count int64,
	value float64,
) {
	var logEvent *zerolog.Event
	if passed {
		logEvent = log.Debug() //nolint:zerologlint
	} else {
		logEvent = log.Error() //nolint:zerologlint
	}

	logEvent.
		Dict("scenario", zerolog.Dict().Str("name", scenario.Name)).
		Str("metric", threshold.Metric).
		Str("type", threshold.Type).
		Float64("rate", rate).
		Int64("count", count).
		Float64("value", value).
		Bool("passed", passed).
		Msg("threshold validation")
}
