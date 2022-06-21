package config

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	App struct {
		Development uint8  `yaml:"development"`
		UseDB       uint8  `yaml:"use_db"`
		GcpID       string `yaml:"-"`
	} `yaml:"app"`
	Binance struct {
		ApiKey     string  `yaml:"-"`
		SecretKey  string  `yaml:"-"`
		Fee        float64 `yaml:"fee"`
		MinProfit  float64 `yaml:"min_profit"`
		Conversion float64 `yaml:"conversion"`
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
	cfg.App.GcpID = os.Getenv("GCP_PROJECT_ID")
	cfg.Binance.ApiKey = os.Getenv("BINANCE_API_KEY")
	cfg.Binance.SecretKey = os.Getenv("BINANCE_SECRET_KEY")
	return cfg
}
