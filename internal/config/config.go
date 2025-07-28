package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env"`
	StoragePath string `yaml:"storage_path" validate:"required"`
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Address string `yaml:"address" env-default:"localhost:8080"`
}

func Load() *Config {
	var cfg Config
	configPath := "config/local.yaml"

	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		log.Fatalf("config file is not exist: %s", configPath)
	}

	err = cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("configuration reading error: %v", err)
	}

	return &cfg
}
