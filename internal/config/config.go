package config

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"time"
)

type Config struct {
	HTTPServer struct {
		ListenAddress string `env:"HTTP_SERVER_LISTEN_ADDRESS,notEmpty" envDefault:":8082"`
	}

	Postgres struct {
		DSN             string        `env:"PG_DSN,notEmpty" envDefault:"postgresql://postgres:postgres@0.0.0.0:5432/postgres?sslmode=disable" json:"-" `
		MaxOpenConns    int           `env:"PG_MAX_OPEN_CONNS" envDefault:"1"`
		ConnMaxLifetime time.Duration `env:"PG_CONN_MAX_LIFETIME" envDefault:"5m"`
	}

	Speller struct {
		URL string `env:"SPELLER_BASE_URL,notEmpty" envDefault:"https://speller.yandex.net/services/spellservice.json"`
	}

	Debug bool `env:"DEBUG" envDefault:"false"`
}

func Load() (Config, error) {
	var config Config

	if err := env.Parse(&config); err != nil {
		return Config{}, fmt.Errorf("env.Parse: %w", err)
	}

	return config, nil
}
