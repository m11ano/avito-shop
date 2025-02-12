package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Fx struct {
		StartTimeout int `yaml:"start_timeout" env:"FX_START_TIMEOUT" env-default:"60"`
		StopTimeout  int `yaml:"stop_timeout" env:"FX_STOP_TIMEOUT" env-default:"10"`
	} `yaml:"fx"`
	DB struct {
		URI                 string `yaml:"uri" env:"DB_URI"`
		MaxAttempt          int    `yaml:"max_attempts" env:"DB_MAX_ATTEMPTS" env-default:"10"`
		AttemptSleepSeconds int    `yaml:"attempt_sleep_seconds" env:"DB_ATTEMPT_SLEEP_SECONDS" env-default:"3"`
	} `yaml:"db"`
}

func LoadConfig() Config {
	//
	var Config Config

	err := cleanenv.ReadConfig("config.yml", &Config)
	if err != nil {
		log.Fatal("config error", err)
	}

	return Config
}
