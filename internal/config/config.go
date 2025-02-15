package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App struct {
		StartTimeout int  `yaml:"start_timeout" env:"APP_START_TIMEOUT" env-default:"60"`
		StopTimeout  int  `yaml:"stop_timeout" env:"APP_STOP_TIMEOUT" env-default:"10"`
		IsProd       bool `yaml:"is_prod" env:"APP_IS_PROD" env-default:"false"`
	} `yaml:"app"`
	DB struct {
		URI                 string `yaml:"uri" env:"DB_URI"`
		MaxAttempt          int    `yaml:"max_attempts" env:"DB_MAX_ATTEMPTS" env-default:"10"`
		AttemptSleepSeconds int    `yaml:"attempt_sleep_seconds" env:"DB_ATTEMPT_SLEEP_SECONDS" env-default:"3"`
	} `yaml:"db"`
	HTTP struct {
		Prefix      string `yaml:"prefix" env:"HTTP_PREFIX" env-default:"api"`
		Port        int    `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
		StopTimeout int    `yaml:"stop_timeout" env:"HTTP_STOP_TIMEOUT" env-default:"5"`
		UnderProxy  bool   `yaml:"under_proxy" env:"HTTP_UNDER_PROXY" env-default:"false"`
	}
	Auth struct {
		JWTSecretKey     string `yaml:"jwt_secret_key" env:"AUTH_JWT_SECRET_KEY" env-default:""`
		JWTTokenTTL      int64  `yaml:"jwt_token_ttl" env:"AUTH_JWT_TOKEN_TTL" env-default:"3600"`
		NewAccountAmount int64  `yaml:"new_account_amount" env:"AUTH_NEW_ACCOUNT_AMOUNT" env-default:"0"`
	}
}

func LoadConfig(file string) Config {
	//
	var Config Config

	err := cleanenv.ReadConfig(file, &Config)
	if err != nil {
		log.Fatal("config error", err)
	}

	return Config
}
