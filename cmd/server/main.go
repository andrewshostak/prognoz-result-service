package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/andrewshostak/result-service/client"
	"github.com/andrewshostak/result-service/config"
	"github.com/andrewshostak/result-service/handler"
	"github.com/andrewshostak/result-service/initializer"
	loggerinternal "github.com/andrewshostak/result-service/logger"
	"github.com/andrewshostak/result-service/middleware"
	"github.com/andrewshostak/result-service/repository"
	"github.com/andrewshostak/result-service/scheduler"
	"github.com/andrewshostak/result-service/service"
	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/procyon-projects/chrono"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "run",
		Short: "Server starts running the server",
		Run:   startServer,
	}

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func startServer(_ *cobra.Command, _ []string) {
	cfg := config.Parse()

	file, err := loggerinternal.GetLogFile()
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func() {
		<-c
		_ = file.Close()
		os.Exit(0)
	}()

	logger := loggerinternal.SetupLogger(file)

	r := gin.Default()

	db := repository.EstablishDatabaseConnection(cfg)
	httpClient := http.Client{}
	chronoTaskScheduler := chrono.NewDefaultTaskScheduler()

	r.Use(middleware.Authorization(cfg.App.HashedAPIKeys, cfg.App.SecretKey))

	v1 := r.Group("/v1")

	footballAPIClient := client.NewFootballAPIClient(&httpClient, logger, cfg.ExternalAPI.FootballAPIBaseURL, cfg.ExternalAPI.RapidAPIKey)
	notifierClient := client.NewNotifierClient(&httpClient, logger)

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
		logger,
		cfg.Result.PollingMaxRetries,
		cfg.Result.PollingInterval,
		cfg.Result.PollingFirstAttemptDelay,
	)
	subscriptionService := service.NewSubscriptionService(subscriptionRepository, matchRepository, aliasRepository, taskScheduler, logger)
	notifierService := service.NewNotifierService(subscriptionRepository, notifierClient, logger)
	aliasService := service.NewAliasService(aliasRepository, logger)

	matchHandler := handler.NewMatchHandler(matchService)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)
	aliasHandler := handler.NewAliasHandler(aliasService)
	v1.POST("/matches", matchHandler.Create)
	v1.POST("/subscriptions", subscriptionHandler.Create)
	v1.DELETE("/subscriptions", subscriptionHandler.Delete)
	v1.GET("/aliases", aliasHandler.Search)

	ctx := context.Background()
	matchResultScheduleInitializer := initializer.NewMatchResultScheduleInitializer(matchService, logger)
	if err := matchResultScheduleInitializer.ReSchedule(ctx); err != nil {
		panic(err)
	}

	notifierInitializer := initializer.NewNotifierInitializer(notifierService)
	notifierInitializer.Start()

	_ = r.Run(fmt.Sprintf(":%s", cfg.App.Port))
}
