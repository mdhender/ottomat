package database

import (
	"context"
	"fmt"
	"log"

	"github.com/mdhender/ottomat/ent"
	_ "modernc.org/sqlite"
)

func Open(dbPath string) (*ent.Client, error) {
	dsn := fmt.Sprintf("file:%s?cache=shared&_fk=1", dbPath)
	client, err := ent.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed opening connection to sqlite: %w", err)
	}
	return client, nil
}

func Migrate(ctx context.Context, client *ent.Client) error {
	if err := client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("failed creating schema resources: %w", err)
	}
	log.Println("database schema created successfully")
	return nil
}
