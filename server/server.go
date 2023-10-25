package server

import (
	"fmt"

	"github.com/andrewshostak/result-service/config"
	"github.com/andrewshostak/result-service/handler"
	"github.com/andrewshostak/result-service/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func StartServer() {
	cfg := config.Parse()

	r := gin.Default()

	_ = establishDatabaseConnection(cfg)

	r.Use(middleware.Authorization(cfg.HashedAPIKeys, cfg.SecretKey))

	v1 := r.Group("/v1")

	matchHandler := handler.NewMatchHandler()
	subscriptionHandler := handler.NewSubscriptionHandler()
	v1.POST("/matches", matchHandler.Create)
	v1.POST("/subscriptions", subscriptionHandler.Create)

	r.Run(fmt.Sprintf(":%s", cfg.Port))
}

func establishDatabaseConnection(cfg config.Server) *gorm.DB {
	connectionParams := fmt.Sprintf(
		"host=%s user=%s password=%s port=%s database=%s sslmode=disable",
		cfg.PGHost,
		cfg.PGUser,
		cfg.PGPassword,
		cfg.PGPort,
		cfg.PGDatabase,
	)

	db, err := gorm.Open(postgres.Open(connectionParams))
	if err != nil {
		panic(err)
	}

	sqlDb, err := db.DB()
	if err != nil {
		panic(err)
	}

	driver, err := migratepg.WithInstance(sqlDb, &migratepg.Config{})
	m, err := migrate.NewWithDatabaseInstance("file://./migrations", cfg.PGDatabase, driver)
	if err != nil {
		panic(err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		panic(err)
	}

	return db
}
