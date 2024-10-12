package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name       string     `yaml:"name"`
	Parallel   bool       `yaml:"parallel"`
	HTTPClient HTTPClient `yaml:"httpClient"`

	DefaultScenario *Scenario   `yaml:"defaults"`
	Scenarios       []*Scenario `yaml:"scenarios"`

	FileName string `yaml:"-"`
}

func (c *Config) ScenarioNames() []string {
	names := make([]string, len(c.Scenarios))

	for i, scenario := range c.Scenarios {
		names[i] = scenario.Name
	}

	return names
}

func Load(fileName string) (*Config, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var conf Config

	d := yaml.NewDecoder(file)
	if err = d.Decode(&conf); err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	conf.FileName = fileName

	conf.applyDefaults()

	return &conf, nil
}

func (c *Config) applyDefaults() {
	for _, scenario := range c.Scenarios {
		MergeScenarios(scenario, c.DefaultScenario)
	}
}
