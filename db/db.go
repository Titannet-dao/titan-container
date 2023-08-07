package db

import (
	"context"
	"embed"
	_ "embed"
	logging "github.com/ipfs/go-log/v2"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var log = logging.Logger("db")

func SqlDB(dsn string) (*sqlx.DB, error) {
	client, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = client.Ping(); err != nil {
		return nil, err
	}

	// initialize the database
	log.Info("db: creating tables")
	err = createAllTables(context.Background(), client)
	if err != nil {
		return nil, errors.Errorf("failed to init db: %v", err)
	}

	return client, nil
}

//go:embed sql/*.sql
var createMainDBSQL embed.FS

func createAllTables(ctx context.Context, mainDB *sqlx.DB) error {
	fileNames := []string{"providers", "deployments", "services", "properties"}

	for _, fileName := range fileNames {
		content, _ := createMainDBSQL.ReadFile("sql/" + fileName + ".sql")
		if _, err := mainDB.ExecContext(ctx, string(content)); err != nil {
			return errors.Errorf("failed to create tables in main DB: %v", err)
		}
	}

	return nil
}
