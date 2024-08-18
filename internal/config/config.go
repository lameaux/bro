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
	MaxIdleConnsPerHost    int           `yaml:"maxIdleConnsPerHost"`
	DisableKeepAlive       bool          `yaml:"disableKeepAlive"`
	Timeout                time.Duration `yaml:"timeout"`
	DisableFollowRedirects bool          `yaml:"disableFollowRedirects"`
}

type Scenario struct {
	Name         string        `yaml:"name"`
	Rate         int           `yaml:"rate"`
	Interval     time.Duration `yaml:"interval"`
	Duration     time.Duration `yaml:"duration"`
	VUs          int           `yaml:"vus"`
	Buffer       int           `yaml:"buffer"`
	PayloadType  string        `yaml:"payloadType"`
	HttpRequest  HttpRequest   `yaml:"httpRequest"`
	HttpResponse HttpResponse  `yaml:"httpResponse"`
	Validate     Validate      `yaml:"validate"`
}

type HttpRequest struct {
	Url    string  `yaml:"url"`
	Method string  `yaml:"method"`
	Body   *string `yaml:"body"`
}

type HttpResponse struct {
	Code int `yaml:"code"`
}

type Validate struct {
	Success *Threshold `yaml:"success"`
	Invalid *Threshold `yaml:"invalid"`
}

type Threshold struct {
	Equal int `yaml:"eq"`
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
