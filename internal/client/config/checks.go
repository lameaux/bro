package config

type Check struct {
	Type     string `yaml:"type"`
	Name     string `yaml:"name"`
	Equals   string `yaml:"equals"`
	Contains string `yaml:"contains"`
	Matches  string `yaml:"matches"`
}
