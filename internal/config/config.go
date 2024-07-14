package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	Name       string     `yaml:"name"`
	Execution  string     `yaml:"execution"`
	HttpClient HttpClient `yaml:"httpClient"`
	Scenarios  []Scenario `yaml:"scenarios"`
}

type HttpClient struct {
	DisableKeepAlive bool          `yaml:"disableKeepAlive"`
	Timeout          time.Duration `yaml:"timeout"`
}

type Scenario struct {
	Name       string        `yaml:"name"`
	Rate       int           `yaml:"rate"`
	Interval   time.Duration `yaml:"interval"`
	VUs        int           `yaml:"vus"`
	Duration   time.Duration `yaml:"duration"`
	Request    HttpRequest   `yaml:"request"`
	Response   HttpResponse  `yaml:"response"`
	Thresholds Thresholds    `yaml:"thresholds"`
}

type HttpRequest struct {
	Url    string  `yaml:"url"`
	Method string  `yaml:"method"`
	Body   *string `yaml:"body"`
}

type HttpResponse struct {
	Code int `yaml:"code"`
}

type Thresholds struct {
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
