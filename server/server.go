package server

import (
	"fmt"
	"github.com/caarlos0/env/v9"
	"github.com/gin-gonic/gin"
)

type config struct {
	Port string `env:"PORT" envDefault:"8080"`
}

func StartServer() {
	config := config{}
	if err := env.Parse(&config); err != nil {
		panic(err)
	}

	r := gin.Default()

	r.Run(fmt.Sprintf(":%s", config.Port))
}
