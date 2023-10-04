package server

import (
	"fmt"

	"github.com/andrewshostak/prognoz-result-service/config"
	"github.com/gin-gonic/gin"
)

func StartServer() {
	cfg := config.Parse()

	r := gin.Default()

	r.Run(fmt.Sprintf(":%s", cfg.Port))
}
