package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/mdhender/ottomat/ent"
	_ "modernc.org/sqlite"
)

func Open(dbPath string) (*ent.Client, error) {
	// hack to prevent Sqlite from creating files when dbPath does not exist
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("database does not exist at %s", dbPath)
	}
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(1000)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed opening database: %w", err)
	}

	// Configure connection pool for SQLite
	db.SetMaxOpenConns(1)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))
	return client, nil
}

func Migrate(ctx context.Context, client *ent.Client) error {
	if err := client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("failed creating schema resources: %w", err)
	}
	log.Println("database schema created successfully")
	return nil
}
