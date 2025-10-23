package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mdhender/ottomat/ent/user"
	"github.com/mdhender/ottomat/internal/database"
	"github.com/mdhender/phrases/v2"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var (
	dbPath              string
	adminPassword       string
	updateAdminPassword string
)

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

		password := adminPassword
		if password == "" {
			password = phrases.Generate(6)
			log.Printf("generated random password: %s", password)
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
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

		log.Printf("created default admin user (username: admin, password: %s)", password)
		return nil
	},
}

var dbUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update database records",
	Long:  `Update existing database records.`,
}

var dbUpdateAdminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Update admin user password",
	Long:  `Update the password for the admin user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := database.Open(dbPath)
		if err != nil {
			return err
		}
		defer client.Close()

		ctx := context.Background()

		adminUser, err := client.User.
			Query().
			Where(user.Username("admin")).
			Only(ctx)
		if err != nil {
			return fmt.Errorf("failed to find admin user: %w", err)
		}

		password := updateAdminPassword
		if password == "" {
			password = phrases.Generate(6)
			log.Printf("generated random password: %s", password)
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		_, err = adminUser.
			Update().
			SetPasswordHash(string(passwordHash)).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update admin password: %w", err)
		}

		log.Printf("updated admin password to: %s", password)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbInitCmd)
	dbCmd.AddCommand(dbMigrateCmd)
	dbCmd.AddCommand(dbSeedCmd)
	dbCmd.AddCommand(dbUpdateCmd)
	dbUpdateCmd.AddCommand(dbUpdateAdminCmd)

	dbCmd.PersistentFlags().StringVar(&dbPath, "db", "ottomat.db", "path to the database file")
	dbSeedCmd.Flags().StringVar(&adminPassword, "password", "", "password for admin user (generates random if not provided)")
	dbUpdateAdminCmd.Flags().StringVar(&updateAdminPassword, "password", "", "new password for admin user (generates random if not provided)")
}
