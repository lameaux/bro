package config

import "time"

type Scenario struct {
	Name string `yaml:"name"`

	HTTPRequest HTTPRequest `yaml:"httpRequest"`

	RpsRaw      int           `yaml:"rps"`
	DurationRaw time.Duration `yaml:"duration"`
	ThreadsRaw  int           `yaml:"threads"`

	Stages []*Stage `yaml:"stages"`

	Checks     []*Check     `yaml:"checks"`
	Thresholds []*Threshold `yaml:"thresholds"`
}

func (s *Scenario) Rps() int {
	return max(s.RpsRaw, 1)
}

func (s *Scenario) Duration() time.Duration {
	return max(s.DurationRaw, 1*time.Second)
}

func (s *Scenario) Threads() int {
	return max(s.ThreadsRaw, 1)
}

func MergeScenarios(scenario *Scenario, defaults *Scenario) *Scenario {
	if defaults == nil {
		return scenario
	}

	scenario.RpsRaw = IntOrDefault(scenario.RpsRaw, defaults.RpsRaw)
	scenario.DurationRaw = DurationOrDefault(scenario.DurationRaw, defaults.DurationRaw)
	scenario.ThreadsRaw = IntOrDefault(scenario.ThreadsRaw, defaults.ThreadsRaw)

	scenario.HTTPRequest.URL = StringOrDefault(scenario.HTTPRequest.URL, defaults.HTTPRequest.URL)
	scenario.HTTPRequest.MethodRaw = PStringOrDefault(scenario.HTTPRequest.MethodRaw, defaults.HTTPRequest.MethodRaw)
	scenario.HTTPRequest.BodyRaw = PStringOrDefault(scenario.HTTPRequest.BodyRaw, defaults.HTTPRequest.BodyRaw)

	scenario.Checks = append(scenario.Checks, defaults.Checks...)
	scenario.Thresholds = append(scenario.Thresholds, defaults.Thresholds...)

	return scenario
}

func IntOrDefault(val int, def int) int {
	if val == 0 {
		return def
	}

	return val
}

func DurationOrDefault(val time.Duration, def time.Duration) time.Duration {
	if val == 0 {
		return def
	}

	return val
}

func StringOrDefault(val string, def string) string {
	if val == "" {
		return def
	}

	return val
}

func PStringOrDefault(val *string, def *string) *string {
	if val == nil {
		return def
	}

	return val
}
