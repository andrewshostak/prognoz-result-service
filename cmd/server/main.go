package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"

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
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "run",
		Short: "Server starts running the server",
		Run:   StartServer,
	}

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func StartServer(_ *cobra.Command, _ []string) {
	cfg := config.Parse()

	file, err := getLogFile()
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func() {
		<-c
		file.Close()
		os.Exit(0)
	}()

	logger := setupLogger(file)

	r := gin.Default()

	db := establishDatabaseConnection(cfg)
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

	matchHandler := handler.NewMatchHandler(matchService)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)
	v1.POST("/matches", matchHandler.Create)
	v1.POST("/subscriptions", subscriptionHandler.Create)
	v1.DELETE("/subscriptions", subscriptionHandler.Delete)

	ctx := context.Background()
	matchResultScheduleInitializer := initializer.NewMatchResultScheduleInitializer(matchService, logger)
	if err := matchResultScheduleInitializer.ReSchedule(ctx); err != nil {
		panic(err)
	}

	notifierInitializer := initializer.NewNotifierInitializer(notifierService)
	notifierInitializer.Start()

	r.Run(fmt.Sprintf(":%s", cfg.App.Port))
}

func establishDatabaseConnection(cfg config.Config) *gorm.DB {
	connectionParams := fmt.Sprintf(
		"host=%s user=%s password=%s port=%s database=%s sslmode=disable",
		cfg.PG.Host,
		cfg.PG.User,
		cfg.PG.Password,
		cfg.PG.Port,
		cfg.PG.Database,
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
	m, err := migrate.NewWithDatabaseInstance("file://./migrations", cfg.PG.Database, driver)
	if err != nil {
		panic(err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		panic(err)
	}

	return db
}

func setupLogger(file io.Writer) *zerolog.Logger {
	logger := zerolog.New(zerolog.MultiLevelWriter(file, os.Stderr)).With().Timestamp().Logger()
	return &logger
}

func getLogFile() (*os.File, error) {
	filename := "app.log"
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s file to write logs: %w", filename, err)
	}

	return file, nil
}
