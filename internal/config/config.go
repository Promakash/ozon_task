package config

import (
	"ozon_task/pkg/infra"
	pkglog "ozon_task/pkg/log"
	"time"
)

type HTTPConfig struct {
	Address           string        `yaml:"address" env-required:"true"`
	ReadTimeout       time.Duration `yaml:"read_timeout" env-default:"5s"`
	WriteTimeout      time.Duration `yaml:"write_timeout" env-default:"5s"`
	IdleTimeout       time.Duration `yaml:"idle_timeout" env-default:"30s"`
	OperationsTimeout time.Duration `yaml:"operations_timeout" env-default:"4s"`
}

type Config struct {
	HTTPServer HTTPConfig           `yaml:"http_server" env-required:"true"`
	GRPC       GRPCConfig           `yaml:"grpc" env-required:"true"`
	PG         infra.PostgresConfig `yaml:"postgres" env-required:"true"`
	Logger     pkglog.Config        `yaml:"logger" env-required:"true"`
}

type GRPCConfig struct {
	Port              int           `yaml:"port" env-required:"true"`
	OperationsTimeout time.Duration `yaml:"operations_timeout" env-default:"5s"`
}
