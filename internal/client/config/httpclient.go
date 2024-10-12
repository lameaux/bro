package config

import "time"

type HTTPClient struct {
	MaxIdleConnsPerHost    int           `yaml:"maxIdleConnsPerHost"`
	DisableKeepAlive       bool          `yaml:"disableKeepAlive"`
	Timeout                time.Duration `yaml:"timeout"`
	DisableFollowRedirects bool          `yaml:"disableFollowRedirects"`
}
