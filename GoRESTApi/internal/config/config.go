package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string     `yaml:"env" env:"ENV" env-required:"true"`
	Storage    string     `yaml:"storage" env:"STORAGE" env-required:"true"`
	HTTPServer HTTPServer `yaml:"http_server"`
}

type HTTPServer struct {
	Addr string `yaml:"address" env:"HTTP_ADDR" env-required:"true"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configFlag := flag.String("config", "", "path to config file")
		flag.Parse()
		configPath = *configFlag
		if configPath == "" {
			log.Fatal("config file path is not specified")
		}
	}

	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("config file error: %s", err)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("failed to read config file: %s", err)
	}

	return &cfg
}
