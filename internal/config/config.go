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
	Name        string        `yaml:"name"`
	RPS         int           `yaml:"rps"`
	Duration    time.Duration `yaml:"duration"`
	VUs         int           `yaml:"vus"`
	Buffer      int           `yaml:"buffer"`
	PayloadType string        `yaml:"payloadType"`
	HttpRequest HttpRequest   `yaml:"httpRequest"`
	Checks      []Check       `yaml:"checks"`
	Thresholds  []Threshold   `yaml:"thresholds"`
}

type HttpRequest struct {
	Url       string  `yaml:"url"`
	MethodRaw string  `yaml:"method"`
	Body      *string `yaml:"body"`
}

func (r HttpRequest) Method() string {
	if r.MethodRaw == "" {
		return http.MethodGet
	}

	return r.MethodRaw
}

func (r HttpRequest) BodyReader() io.Reader {
	if r.Body != nil {
		return strings.NewReader(*r.Body)
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
	Type string `yaml:"type"`
	Name string `yaml:"name"`

	MinCount int `yaml:"minCount"`
	MaxCount int `yaml:"maxCount"`

	MinValue string `yaml:"minValue"`
	MaxValue string `yaml:"maxValue"`

	MinRate string `yaml:"minRate"`
	MaxRate string `yaml:"maxRate"`
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
