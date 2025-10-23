package main

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
	dbPath         string
	adminUsername  string
	adminPassword  string
	createPassword string
	createRole     string
	createClanID   int
	updatePassword string
	updateRole     string
	updateClanID   int
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

		username := adminUsername
		if username == "" {
			username = "admin"
		}

		exists, err := client.User.
			Query().
			Where(user.Username(username)).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed to check for admin user: %w", err)
		}
		if exists {
			log.Printf("admin user '%s' already exists", username)
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
			SetUsername(username).
			SetPasswordHash(string(passwordHash)).
			SetRole(user.RoleAdmin).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		log.Printf("created admin user (username: %s, password: %s)", username, password)
		return nil
	},
}

var dbCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create database records",
	Long:  `Create new database records.`,
}

var dbCreateUserCmd = &cobra.Command{
	Use:   "user <username>",
	Short: "Create a new user",
	Long:  `Create a new user with specified username and optional password, role, and clan ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		username := args[0]

		client, err := database.Open(dbPath)
		if err != nil {
			return err
		}
		defer client.Close()

		ctx := context.Background()

		// Check if user already exists
		exists, err := client.User.
			Query().
			Where(user.Username(username)).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed to check for existing user: %w", err)
		}
		if exists {
			return fmt.Errorf("user '%s' already exists", username)
		}

		password := createPassword
		if password == "" {
			password = phrases.Generate(6)
			log.Printf("generated random password: %s", password)
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		role := createRole
		if role == "" {
			role = "guest"
		}

		create := client.User.
			Create().
			SetUsername(username).
			SetPasswordHash(string(passwordHash)).
			SetRole(user.Role(role))

		if createClanID != 0 {
			create.SetClanID(createClanID)
		}

		newUser, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		clanInfo := "none"
		if newUser.ClanID != nil {
			clanInfo = fmt.Sprintf("%d", *newUser.ClanID)
		}

		log.Printf("created user '%s' (role: %s, clan: %s, password: %s)", username, role, clanInfo, password)
		return nil
	},
}

var dbUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update database records",
	Long:  `Update existing database records.`,
}

var dbUpdateUserCmd = &cobra.Command{
	Use:   "user <username>",
	Short: "Update user record",
	Long:  `Update fields for a specific user. At least one update flag must be provided.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		username := args[0]

		// Check that at least one update flag was provided
		passwordSet := cmd.Flags().Changed("password")
		roleSet := cmd.Flags().Changed("role")
		clanIDSet := cmd.Flags().Changed("clan-id")

		if !passwordSet && !roleSet && !clanIDSet {
			return fmt.Errorf("at least one update flag must be provided (--password, --role, --clan-id)")
		}

		client, err := database.Open(dbPath)
		if err != nil {
			return err
		}
		defer client.Close()

		ctx := context.Background()

		targetUser, err := client.User.
			Query().
			Where(user.Username(username)).
			Only(ctx)
		if err != nil {
			return fmt.Errorf("failed to find user '%s': %w", username, err)
		}

		update := targetUser.Update()
		updates := []string{}

		if passwordSet {
			password := updatePassword
			if password == "" {
				password = phrases.Generate(6)
				log.Printf("generated random password: %s", password)
			}

			passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("failed to hash password: %w", err)
			}

			update.SetPasswordHash(string(passwordHash))
			updates = append(updates, fmt.Sprintf("password: %s", password))
		}

		if roleSet {
			update.SetRole(user.Role(updateRole))
			updates = append(updates, fmt.Sprintf("role: %s", updateRole))
		}

		if clanIDSet {
			if updateClanID == 0 {
				update.ClearClanID()
				updates = append(updates, "clan_id: cleared")
			} else {
				update.SetClanID(updateClanID)
				updates = append(updates, fmt.Sprintf("clan_id: %d", updateClanID))
			}
		}

		_, err = update.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		log.Printf("updated user '%s': %v", username, updates)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbInitCmd)
	dbCmd.AddCommand(dbMigrateCmd)
	dbCmd.AddCommand(dbSeedCmd)
	dbCmd.AddCommand(dbCreateCmd)
	dbCmd.AddCommand(dbUpdateCmd)
	dbCreateCmd.AddCommand(dbCreateUserCmd)
	dbUpdateCmd.AddCommand(dbUpdateUserCmd)

	dbCmd.PersistentFlags().StringVar(&dbPath, "db", "ottomat.db", "path to the database file")
	dbSeedCmd.Flags().StringVar(&adminUsername, "username", "admin", "username for admin user")
	dbSeedCmd.Flags().StringVar(&adminPassword, "password", "", "password for admin user (generates random if not provided)")
	dbCreateUserCmd.Flags().StringVar(&createPassword, "password", "", "password for user (generates random if not provided)")
	dbCreateUserCmd.Flags().StringVar(&createRole, "role", "guest", "role for user (guest, chief, admin)")
	dbCreateUserCmd.Flags().IntVar(&createClanID, "clan-id", 0, "clan ID for user")
	dbUpdateUserCmd.Flags().StringVar(&updatePassword, "password", "", "new password for user (generates random if not provided)")
	dbUpdateUserCmd.Flags().StringVar(&updateRole, "role", "", "new role for user (guest, chief, admin)")
	dbUpdateUserCmd.Flags().IntVar(&updateClanID, "clan-id", 0, "new clan ID for user (0 to clear)")
}
