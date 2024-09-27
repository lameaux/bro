package config

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name       string      `yaml:"name"`
	Parallel   bool        `yaml:"parallel"`
	HTTPClient HTTPClient  `yaml:"httpClient"`
	Scenarios  []*Scenario `yaml:"scenarios"`

	FileName string `yaml:"-"`
}

type HTTPClient struct {
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
	HTTPRequest HTTPRequest `yaml:"httpRequest"`
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

type HTTPRequest struct {
	URL       string  `yaml:"url"`
	MethodRaw *string `yaml:"method"`
	BodyRaw   *string `yaml:"body"`
}

func (r *HTTPRequest) Method() string {
	if r.MethodRaw != nil {
		return *r.MethodRaw
	}

	return http.MethodGet
}

func (r *HTTPRequest) BodyReader() io.Reader {
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

	return &conf, nil
}
