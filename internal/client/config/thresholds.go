package config

type Threshold struct {
	Metric string `yaml:"metric"`
	Type   string `yaml:"type"`

	MinCount *int64 `yaml:"minCount"`
	MaxCount *int64 `yaml:"maxCount"`

	MinValue *float64 `yaml:"minValue"`
	MaxValue *float64 `yaml:"maxValue"`

	MinRate *float64 `yaml:"minRate"`
	MaxRate *float64 `yaml:"maxRate"`
}
