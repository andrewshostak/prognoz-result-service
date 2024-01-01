package main

import (
	"context"
	"net/http"

	"github.com/andrewshostak/result-service/client"
	"github.com/andrewshostak/result-service/config"
	"github.com/andrewshostak/result-service/helper"
	loggerinternal "github.com/andrewshostak/result-service/logger"
	"github.com/andrewshostak/result-service/repository"
	"github.com/andrewshostak/result-service/service"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "run",
		Short: "Backfills aliases",
		Run:   run,
	}

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func run(_ *cobra.Command, _ []string) {
	cfg := config.Parse()

	file, err := loggerinternal.GetLogFile()
	if err != nil {
		panic(err)
	}

	logger := loggerinternal.SetupLogger(file)

	httpClient := http.Client{}

	db := repository.EstablishDatabaseConnection(cfg)

	aliasRepository := repository.NewAliasRepository(db)

	seasonHelper := helper.NewSeasonHelper()

	footballAPIClient := client.NewFootballAPIClient(&httpClient, logger, cfg.ExternalAPI.FootballAPIBaseURL, cfg.ExternalAPI.RapidAPIKey)

	backfillAliasesService := service.NewBackfillAliasesService(aliasRepository, footballAPIClient, seasonHelper, logger)

	ctx := context.Background()

	err = backfillAliasesService.Backfill(ctx)
	if err != nil {
		panic(err)
	}
}
