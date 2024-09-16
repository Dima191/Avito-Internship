package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Address string `env:"SERVER_ADDRESS"`
	DB      struct {
		PostgresConnStr  string `env:"POSTGRES_CONN"`
		PostgresUserName string `env:"POSTGRES_USERNAME"`
		PostgresPassword string `env:"POSTGRES_PASSWORD"`
		PostgresHost     string `env:"POSTGRES_HOST"`
		PostgresPort     string `env:"POSTGRES_PORT"`
		PostgresDatabase string `env:"POSTGRES_DATABASE"`
	}
}

func New() (*Config, error) {
	cfg := new(Config)
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
