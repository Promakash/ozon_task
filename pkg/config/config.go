package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

func MustLoad[T any](configEnv string) *T {
	configPath := os.Getenv(configEnv)
	if configPath == "" {
		log.Fatalf("Env variable: %s is not set", configEnv)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg T

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("error reading config: %s", err)
	}

	return &cfg
}
