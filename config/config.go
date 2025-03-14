package config

import (
	"fmt"

	"github.com/caarlos0/env/v9"
	"github.com/go-playground/validator/v10"
)

type Config struct {
	APP_PORT           string `env:"APP_PORT" envDefault:"8085" validate:"required,numeric,gte=1"`
	APP_MAX_READ_TIME  int    `env:"APP_MAX_READ_TIME" envDefault:"60" validate:"required,gte=1"`
	APP_MAX_WRITE_TIME int    `env:"APP_MAX_WRITE_TIME" envDefault:"60" validate:"required,gte=1"`
}

var Conf Config
var validate *validator.Validate

func init() {
	validate = validator.New()
}

func LoadEnv() (*Config, error) {
	if err := env.Parse(&Conf); err != nil {
		return nil, fmt.Errorf("failed to load the env: %v", err)
	}
	if err := validate.Struct(Conf); err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}
	return &Conf, nil
}
