package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env       string `yaml:"env" env:"ENV" env-default:"local"`
	AppSecret string `yaml:"app_secret" env:"APP_SECRET" env-required:"true"`

	Storage struct {
		Path string `yaml:"path" env:"STORAGE_PATH" env-required:"true"`
	} `yaml:"storage"`

	GRPC struct {
		Server struct {
			Port    int           `yaml:"port" env:"GRPC_PORT"`
			Timeout time.Duration `yaml:"timeout" env:"GRPC_TIMEOUT"`
		} `yaml:"server"`
	} `yaml:"grpc"`

	Clients struct {
		CRUD struct {
			Addr         string        `yaml:"addr" env:"CRUD_ADDR"`
			Timeout      time.Duration `yaml:"timeout" env:"CRUD_TIMEOUT"`
			RetriesCount int           `yaml:"retries_count" env:"CRUD_RETRIES_COUNT"`
		} `yaml:"crud"`
	} `yaml:"clients"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH_FOR_CRUD")
	if configPath == "" {
		panic("CONFIG_PATH environment variable not set")
	}
	var config Config
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		panic("failed to read config: " + err.Error())
	}
	return &config
}
