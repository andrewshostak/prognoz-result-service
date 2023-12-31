package main

import (
	"fmt"

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
	fmt.Println("backfilling aliases")
}
