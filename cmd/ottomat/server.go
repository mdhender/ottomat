package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mdhender/ottomat/internal/database"
	"github.com/mdhender/ottomat/internal/server"
	"github.com/spf13/cobra"
)

var (
	serverPort    string
	serverTimeout time.Duration
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the web server",
	Long:  `Start the OttoMat web server with graceful shutdown support.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := database.Open(dbPath)
		if err != nil {
			return err
		}
		defer client.Close()

		srv := server.New(client)
		httpServer := &http.Server{
			Addr:    ":" + serverPort,
			Handler: srv,
		}

		serverErrors := make(chan error, 1)
		go func() {
			log.Printf("server listening on port %s\n", serverPort)
			serverErrors <- httpServer.ListenAndServe()
		}()

		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

		if serverTimeout > 0 {
			go func() {
				time.Sleep(serverTimeout)
				log.Printf("timeout reached (%v), initiating shutdown\n", serverTimeout)
				shutdown <- syscall.SIGTERM
			}()
		}

		select {
		case err := <-serverErrors:
			return fmt.Errorf("server error: %w", err)
		case sig := <-shutdown:
			log.Printf("received signal %v, starting graceful shutdown\n", sig)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := httpServer.Shutdown(ctx); err != nil {
				log.Printf("error during shutdown: %v\n", err)
				return httpServer.Close()
			}

			log.Println("server stopped gracefully")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&serverPort, "port", "8080", "port to listen on")
	serverCmd.Flags().DurationVar(&serverTimeout, "timeout", 0, "automatically shutdown after duration (for testing)")
	serverCmd.Flags().StringVar(&dbPath, "db", "ottomat.db", "path to the database file")
}
