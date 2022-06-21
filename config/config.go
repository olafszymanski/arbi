package config

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	App struct {
		Development uint8  `yaml:"development"`
		UseDB       uint8  `yaml:"use_db"`
		GcpID       string `yaml:"gcp_id"`
	} `yaml:"app"`
	Binance struct {
		WebsocketScheme string  `yaml:"websocket_scheme"`
		WebsocketHost   string  `yaml:"websocket_host"`
		ApiScheme       string  `yaml:"api_scheme"`
		ApiHost         string  `yaml:"api_host"`
		ApiKey          string  `yaml:"api_key"`
		SecretKey       string  `yaml:"secret_key"`
		Fee             float64 `yaml:"fee"`
		MinProfit       float64 `yaml:"min_profit"`
		Conversion      float64 `yaml:"conversion"`
	} `yaml:"binance"`
}

func NewConfig(path string) *Config {
	var cfg *Config
	f, err := ioutil.ReadFile(path)
	if err != nil {
		log.WithError(err).Panic()
	}
	if err := yaml.Unmarshal(f, &cfg); err != nil {
		log.WithError(err).Panic()
	}
	return cfg
}
