package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mdhender/ottomat/ent/user"
	"github.com/mdhender/ottomat/internal/database"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var dbPath string

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
	Long:  `Manage the OttoMat database including migrations and seeding.`,
}

var dbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	Long:  `Create the database file if it doesn't exist.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(dbPath); err == nil {
			return fmt.Errorf("database already exists at %s", dbPath)
		}
		file, err := os.Create(dbPath)
		if err != nil {
			return fmt.Errorf("failed to create database file: %w", err)
		}
		file.Close()
		log.Printf("database initialized at %s\n", dbPath)
		return nil
	},
}

var dbMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Apply schema migrations to the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := database.Open(dbPath)
		if err != nil {
			return err
		}
		defer client.Close()

		ctx := context.Background()
		if err := database.Migrate(ctx, client); err != nil {
			return err
		}
		log.Println("migrations completed successfully")
		return nil
	},
}

var dbSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed the database with initial data",
	Long:  `Create default admin user and other initial data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := database.Open(dbPath)
		if err != nil {
			return err
		}
		defer client.Close()

		ctx := context.Background()

		exists, err := client.User.
			Query().
			Where(user.Username("admin")).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed to check for admin user: %w", err)
		}
		if exists {
			log.Println("admin user already exists")
			return nil
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		_, err = client.User.
			Create().
			SetUsername("admin").
			SetPasswordHash(string(passwordHash)).
			SetRole(user.RoleAdmin).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		log.Println("created default admin user (username: admin, password: admin)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbInitCmd)
	dbCmd.AddCommand(dbMigrateCmd)
	dbCmd.AddCommand(dbSeedCmd)

	dbCmd.PersistentFlags().StringVar(&dbPath, "db", "ottomat.db", "path to the database file")
}
