package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ottomat",
	Short: "OttoMat web server",
	Long:  `OttoMat is a web server with HTMX frontend and Go backend.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
