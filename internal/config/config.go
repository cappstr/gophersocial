package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type ENV string

const (
	dev  ENV = "dev"
	prod ENV = "prod"
)

type Config struct {
	ApiServerHost    string `env:"API_SERVER_HOST"`
	ApiServerAddr    string `env:"API_SERVER_ADDR"`
	DatabaseName     string `env:"DB_NAME"`
	DatabaseHost     string `env:"DB_HOST"`
	DatabasePort     string `env:"DB_PORT"`
	DatabaseUser     string `env:"DB_USER"`
	DatabasePassword string `env:"DB_PASSWORD"`
	Env              ENV    `env:"ENV" envDefault:"prod"`
	DatabaseTestPort string `env:"DB_TEST_PORT"`
	JwtSecret        string `env:"JWT_SECRET"`
}

func (c *Config) DatabaseUrl() string {
	databasePort := c.DatabasePort
	if c.Env == dev {
		databasePort = c.DatabaseTestPort
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", c.DatabaseUser, c.DatabasePassword,
		c.DatabaseHost, databasePort, c.DatabaseName)
}

func NewConfig() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}
	return &cfg, nil
}
