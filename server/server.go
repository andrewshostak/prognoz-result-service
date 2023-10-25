package server

import (
	"fmt"

	"github.com/andrewshostak/result-service/config"
	"github.com/andrewshostak/result-service/handler"
	"github.com/andrewshostak/result-service/middleware"
	"github.com/gin-gonic/gin"
)

func StartServer() {
	cfg := config.Parse()

	r := gin.Default()

	r.Use(middleware.Authorization(cfg.HashedAPIKeys, cfg.SecretKey))

	v1 := r.Group("/v1")

	matchHandler := handler.NewMatchHandler()
	subscriptionHandler := handler.NewSubscriptionHandler()
	v1.POST("/matches", matchHandler.Create)
	v1.POST("/subscriptions", subscriptionHandler.Create)

	r.Run(fmt.Sprintf(":%s", cfg.Port))
}
