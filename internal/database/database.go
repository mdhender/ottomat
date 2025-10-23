package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/mdhender/ottomat/ent"
	_ "modernc.org/sqlite"
)

func Open(dbPath string) (*ent.Client, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed opening database: %w", err)
	}
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
