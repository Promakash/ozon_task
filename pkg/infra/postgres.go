package infra

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"db_name"`
}

func NewPostgresPool(cfg PostgresConfig) (*pgxpool.Pool, error) {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)

	dbPool, err := pgxpool.New(context.Background(), psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("can't create connection to postgres: %w", err)
	}

	return dbPool, nil
}
