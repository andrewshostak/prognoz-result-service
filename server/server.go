package server

import (
	"fmt"

	"github.com/andrewshostak/prognoz-result-service/config"
	"github.com/andrewshostak/prognoz-result-service/handler"
	"github.com/andrewshostak/prognoz-result-service/middleware"
	"github.com/gin-gonic/gin"
)

func StartServer() {
	cfg := config.Parse()

	r := gin.Default()

	r.Use(middleware.Authorization(cfg.HashedAPIKeys, cfg.SecretKey))

	v1 := r.Group("/v1")

	subscriptionHandler := handler.NewSubscriptionHandler()
	v1.POST("/subscriptions", subscriptionHandler.Create)

	r.Run(fmt.Sprintf(":%s", cfg.Port))
}
