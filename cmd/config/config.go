package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Binance struct {
		WebsocketScheme string  `yaml:"websocket_scheme"`
		WebsocketHost   string  `yaml:"websocket_host"`
		ApiScheme       string  `yaml:"api_scheme"`
		ApiHost         string  `yaml:"api_host"`
		ApiKey          string  `yaml:"api_key"`
		SecretKey       string  `yaml:"secret_key"`
		Fee             float64 `yaml:"fee"`
		MinProfit       float64 `yaml:"min_profit"`
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
