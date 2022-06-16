package config

import (
	"io/ioutil"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

type Config struct {
	App struct {
		UseDB int8 `yaml:"use_db"`
	} `yaml:"app"`
	Database struct {
		Host     string `yaml:"host"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
		Driver   string `yaml:"driver"`
	} `yaml:"database"`
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

func NewConfig(l *zerolog.Logger) *Config {
	var cfg *Config
	f, err := ioutil.ReadFile("config/config.yml")
	if err != nil {
		l.Panic().Msg(err.Error())
	}
	if err := yaml.Unmarshal(f, &cfg); err != nil {
		l.Panic().Msg(err.Error())
	}
	return cfg
}
