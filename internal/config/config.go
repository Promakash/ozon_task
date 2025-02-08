package config

import (
	"ozon_task/pkg/infra"
	pkglog "ozon_task/pkg/log"
	"time"
)

type HTTPConfig struct {
	Address           string        `yaml:"address"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
	OperationsTimeout time.Duration `yaml:"operations_timeout"`
}

type Config struct {
	HTTPServer HTTPConfig           `yaml:"http_server"`
	GRPC       GRPCConfig           `yaml:"grpc"`
	PG         infra.PostgresConfig `yaml:"postgres"`
	Logger     pkglog.Config        `yaml:"logger"`
}

type GRPCConfig struct {
	Port              int           `yaml:"port"`
	OperationsTimeout time.Duration `yaml:"operations_timeout"`
}
