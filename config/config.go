package config

import "github.com/caarlos0/env/v9"

type Server struct {
	Port                string   `env:"PORT" envDefault:"8080"`
	RapidAPIKey         string   `env:"RAPID_API_KEY,required"`
	FootballAPITimezone string   `env:"FOOTBALL_API_TIMEZONE" envDefault:"Europe/Kiev"`
	HashedAPIKeys       []string `env:"HASHED_API_KEYS" envSeparator:","`
	SecretKey           string   `env:"SECRET_KEY,required"`
}

func Parse() Server {
	config := Server{}
	if err := env.Parse(&config); err != nil {
		panic(err)
	}

	return config
}
