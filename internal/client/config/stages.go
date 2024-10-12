package config

import "time"

type Stage struct {
	Name        string        `yaml:"name"`
	RpsRaw      int           `yaml:"rps"`
	DurationRaw time.Duration `yaml:"duration"`
	ThreadsRaw  int           `yaml:"threads"`
}

func (s *Stage) Rps() int {
	return max(s.RpsRaw, 1)
}

func (s *Stage) Duration() time.Duration {
	return max(s.DurationRaw, 1*time.Second)
}

func (s *Stage) Threads() int {
	return max(s.ThreadsRaw, 1)
}
