package config

import (
	"time"

	"github.com/caarlos0/env/v9"
)

type Server struct {
	Port                string   `env:"PORT" envDefault:"8080"`
	RapidAPIKey         string   `env:"RAPID_API_KEY,required"`
	FootballAPITimezone string   `env:"FOOTBALL_API_TIMEZONE" envDefault:"Europe/Kiev"`
	FootballAPIBaseURL  string   `env:"FOOTBALL_API_BASE_URL" envDefault:"https://api-football-v1.p.rapidapi.com"`
	HashedAPIKeys       []string `env:"HASHED_API_KEYS" envSeparator:","`
	SecretKey           string   `env:"SECRET_KEY,required"`
	PGHost              string   `env:"PG_HOST" envDefault:"localhost"`
	PGUser              string   `env:"PG_USER" envDefault:"postgres"`
	PGPassword          string   `env:"PG_PASSWORD,required"`
	PGPort              string   `env:"PG_PORT" envDefault:"5432"`
	PGDatabase          string   `env:"PG_DATABASE" envDefault:"postgres"`
}

func (s *Server) Location() *time.Location {
	location, _ := time.LoadLocation(s.FootballAPITimezone)
	return location
}

func Parse() Server {
	config := Server{}
	if err := env.Parse(&config); err != nil {
		panic(err)
	}

	if err := validateTimezone(config.FootballAPITimezone); err != nil {
		panic(err)
	}

	return config
}

func validateTimezone(timezone string) error {
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return err
	}

	return nil
}
