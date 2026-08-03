package config

import (
	"os"

	"github.com/goccy/go-yaml"
)

type AllowedRule struct {
	Host       string   `yaml:"host"`
	Hostname   string   `yaml:"hostname"`
	KeyPath    string   `yaml:"key_path"`
	PathPrefix []string `yaml:"path_prefix"`
}

type Config struct {
	Allowed []AllowedRule `yaml:"allowed"`
	LogPath string        `yaml:"logpath"`
}

func ReadConfig(path string) (*Config, error) {
	configData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	conf := &Config{
		LogPath: "/tmp/ssh-wrapper.log",
	}
	err = yaml.Unmarshal(configData, conf)
	if conf.LogPath == "" {
		conf.LogPath = "/tmp/ssh-wrapper.log"
	}
	return conf, err
}
