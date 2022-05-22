package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Binance struct {
		ApiKey    string `yaml:"api_key"`
		SecretKey string `yaml:"secret_key"`
	} `yaml:"binance"`
}

func NewConfig() *Config {
	var cfg *Config
	f, err := ioutil.ReadFile("config/config.yml")
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(f, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
