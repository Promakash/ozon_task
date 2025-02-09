package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

func MustLoad(configEnv string, cfg interface{}) {
	configPath := fetchConfigPath(configEnv)
	if configPath == "" {
		log.Fatal("Config path is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist by this path: %s", configPath)
	}

	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		log.Fatalf("error reading config: %s", err)
	}
}

func fetchConfigPath(configEnv string) string {
	var path string

	flag.StringVar(&path, "config", "", "path to config file")
	flag.Parse()

	if path == "" {
		path = os.Getenv(configEnv)
	}

	return path
}
