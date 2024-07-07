package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	Name      string     `yaml:"name"`
	Execution string     `yaml:"execution"`
	Scenarios []Scenario `yaml:"scenarios"`
}

type Scenario struct {
	Name     string        `yaml:"name"`
	Rate     int           `yaml:"rate"`
	Interval time.Duration `yaml:"interval"`
	VUs      int           `yaml:"vus"`
	Duration time.Duration `yaml:"duration"`
	Request  Request       `yaml:"request"`
}

type Request struct {
	Http HttpRequest `yaml:"http"`
}

type HttpRequest struct {
	Url              string        `yaml:"url"`
	Method           string        `yaml:"method"`
	Body             *string       `yaml:"body"`
	DisableKeepAlive bool          `yaml:"disableKeepAlive"`
	Timeout          time.Duration `yaml:"timeout"`
}

func LoadFromFile(fileName string) (*Config, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var c Config
	d := yaml.NewDecoder(file)
	if err = d.Decode(&c); err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	return &c, nil
}
