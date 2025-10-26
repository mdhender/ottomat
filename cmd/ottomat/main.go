package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	log.SetFlags(log.Lshortfile)

	var rootCmd = &cobra.Command{
		Use:   "ottomat",
		Short: "OttoMat web server",
		Long:  `OttoMat is a web server with HTMX frontend and Go backend.`,
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(cmdDb)
	cmdDb.AddCommand(cmdDbCreate)
	cmdDb.AddCommand(cmdDbInit)
	cmdDb.AddCommand(cmdDbMigrate)
	cmdDb.AddCommand(cmdDbSeed)
	cmdDb.AddCommand(cmdDbUpdate)
	cmdDbCreate.AddCommand(cmdDbCreateUser)
	cmdDbUpdate.AddCommand(cmdDbUpdateUser)

	cmdDb.PersistentFlags().StringVar(&dbPath, "db", "ottomat.db", "path to the database file")
	cmdDbCreateUser.Flags().IntVar(&createClanID, "clan-id", 0, "clan ID for user")
	cmdDbCreateUser.Flags().StringVar(&createPassword, "password", "", "password for user (generates random if not provided)")
	cmdDbCreateUser.Flags().StringVar(&createRole, "role", "guest", "role for user (guest, chief, admin)")
	cmdDbSeed.Flags().StringVar(&adminPassword, "password", "", "password for admin user (generates random if not provided)")
	cmdDbSeed.Flags().StringVar(&adminUsername, "username", "admin", "username for admin user")
	cmdDbUpdateUser.Flags().IntVar(&updateClanID, "clan-id", 0, "new clan ID for user (0 to clear)")
	cmdDbUpdateUser.Flags().StringVar(&updatePassword, "password", "", "new password for user (generates random if not provided)")
	cmdDbUpdateUser.Flags().StringVar(&updateRole, "role", "", "new role for user (guest, chief, admin)")

	rootCmd.AddCommand(cmdServer)
	cmdServer.Flags().BoolVar(&devMode, "dev", false, "enable development mode (disables password managers)")
	cmdServer.Flags().BoolVar(&visiblePasswords, "visible-passwords", false, "show passwords as plain text (requires --dev)")
	cmdServer.Flags().DurationVar(&serverTimeout, "timeout", 0, "automatically shutdown after duration (for testing)")
	cmdServer.Flags().StringVar(&dbPath, "db", "./ottomat.db", "path to the database file")
	cmdServer.Flags().StringVar(&serverPort, "port", "8080", "port to listen on")

	rootCmd.AddCommand(cmdVersion)
	cmdVersion.Flags().BoolVar(&buildInfo, "build-info", false, "show build information")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
