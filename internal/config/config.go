package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	Name       string      `yaml:"name"`
	Execution  string      `yaml:"execution"`
	HttpClient HttpClient  `yaml:"httpClient"`
	Scenarios  []*Scenario `yaml:"scenarios"`
}

type HttpClient struct {
	MaxIdleConnsPerHost    int           `yaml:"maxIdleConnsPerHost"`
	DisableKeepAlive       bool          `yaml:"disableKeepAlive"`
	Timeout                time.Duration `yaml:"timeout"`
	DisableFollowRedirects bool          `yaml:"disableFollowRedirects"`
}

type Scenario struct {
	Name string `yaml:"name"`

	RpsRaw       int           `yaml:"rps"`
	DurationRaw  time.Duration `yaml:"duration"`
	ThreadsRaw   int           `yaml:"threads"`
	QueueSizeRaw int           `yaml:"queueSize"`

	PayloadType string      `yaml:"payloadType"`
	HttpRequest HttpRequest `yaml:"httpRequest"`
	Checks      []Check     `yaml:"checks"`
	Thresholds  []Threshold `yaml:"thresholds"`
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

func (s *Scenario) QueueSize() int {
	return max(s.QueueSizeRaw, s.Threads())
}

type HttpRequest struct {
	Url       string  `yaml:"url"`
	MethodRaw *string `yaml:"method"`
	BodyRaw   *string `yaml:"body"`
}

func (r *HttpRequest) Method() string {
	if r.MethodRaw != nil {
		return *r.MethodRaw

	}

	return http.MethodGet
}

func (r *HttpRequest) BodyReader() io.Reader {
	if r.BodyRaw != nil {
		return strings.NewReader(*r.BodyRaw)
	}

	return nil
}

type Check struct {
	Type     string `yaml:"type"`
	Name     string `yaml:"name"`
	Equals   string `yaml:"equals"`
	Contains string `yaml:"contains"`
	Matches  string `yaml:"matches"`
}

type Threshold struct {
	Metric string `yaml:"metric"`
	Type   string `yaml:"type"`

	MinCount *int `yaml:"minCount"`
	MaxCount *int `yaml:"maxCount"`

	MinValue *int `yaml:"minValue"`
	MaxValue *int `yaml:"maxValue"`

	MinRate *float64 `yaml:"minRate"`
	MaxRate *float64 `yaml:"maxRate"`
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
