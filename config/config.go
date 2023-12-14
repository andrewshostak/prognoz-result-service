package config

import "github.com/caarlos0/env/v9"

type Server struct {
	Port               string   `env:"PORT" envDefault:"8080"`
	RapidAPIKey        string   `env:"RAPID_API_KEY,required"`
	FootballAPIBaseURL string   `env:"FOOTBALL_API_BASE_URL" envDefault:"https://api-football-v1.p.rapidapi.com"`
	HashedAPIKeys      []string `env:"HASHED_API_KEYS" envSeparator:","`
	SecretKey          string   `env:"SECRET_KEY,required"`
	PGHost             string   `env:"PG_HOST" envDefault:"localhost"`
	PGUser             string   `env:"PG_USER" envDefault:"postgres"`
	PGPassword         string   `env:"PG_PASSWORD,required"`
	PGPort             string   `env:"PG_PORT" envDefault:"5432"`
	PGDatabase         string   `env:"PG_DATABASE" envDefault:"postgres"`
}

func Parse() Server {
	config := Server{}
	if err := env.Parse(&config); err != nil {
		panic(err)
	}

	return config
}
