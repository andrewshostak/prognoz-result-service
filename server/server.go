package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/andrewshostak/result-service/client"
	"github.com/andrewshostak/result-service/config"
	"github.com/andrewshostak/result-service/handler"
	"github.com/andrewshostak/result-service/initializer"
	"github.com/andrewshostak/result-service/middleware"
	"github.com/andrewshostak/result-service/repository"
	"github.com/andrewshostak/result-service/scheduler"
	"github.com/andrewshostak/result-service/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/procyon-projects/chrono"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func StartServer() {
	cfg := config.Parse()

	r := gin.Default()

	db := establishDatabaseConnection(cfg)
	httpClient := http.Client{}
	chronoTaskScheduler := chrono.NewDefaultTaskScheduler()

	r.Use(middleware.Authorization(cfg.HashedAPIKeys, cfg.SecretKey))

	v1 := r.Group("/v1")

	footballAPIClient := client.NewFootballAPIClient(&httpClient, cfg.FootballAPIBaseURL, cfg.RapidAPIKey)

	aliasRepository := repository.NewAliasRepository(db)
	matchRepository := repository.NewMatchRepository(db)
	footballAPIFixtureRepository := repository.NewFootballAPIFixtureRepository(db)
	subscriptionRepository := repository.NewSubscriptionRepository(db)

	taskScheduler := scheduler.NewTaskScheduler(chronoTaskScheduler)

	matchService := service.NewMatchService(
		aliasRepository,
		matchRepository,
		footballAPIFixtureRepository,
		footballAPIClient,
		taskScheduler,
		cfg.Location(),
	)
	subscriptionService := service.NewSubscriptionService(subscriptionRepository, matchRepository)

	matchHandler := handler.NewMatchHandler(matchService)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)
	v1.POST("/matches", matchHandler.Create)
	v1.POST("/subscriptions", subscriptionHandler.Create)

	ctx := context.Background()
	matchResultScheduleInitializer := initializer.NewMatchResultScheduleInitializer(matchService)
	if err := matchResultScheduleInitializer.ReSchedule(ctx); err != nil {
		panic(err)
	}

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
