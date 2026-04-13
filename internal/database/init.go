package database

import (
	"fmt"
	"os"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// QB is the global query builder, pre-configured with SQLite's ? placeholder format.
var QB = sq.StatementBuilder.PlaceholderFormat(sq.Question)

// New opens the SQLite database at path, applies recommended pragmas, and
// initialises the schema from schemaPath. The caller is responsible for
// closing the returned *sqlx.DB.
func New(path, schemaPath string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open the database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	pragmas := []string{
		// Enable referential integrity checks.
		"PRAGMA foreign_keys = ON",
		// WAL mode allows concurrent readers alongside a single writer,
		// which improves throughput under typical web-server workloads.
		"PRAGMA journal_mode = WAL",
		// Wait up to 5 s when the database is locked by another writer
		// before returning SQLITE_BUSY, avoiding spurious errors under
		// brief write contention.
		"PRAGMA busy_timeout = 5000",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("failed to set pragma (%s): %w", pragma, err)
		}
	}

	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	if _, err := db.Exec(string(schemaBytes)); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}
