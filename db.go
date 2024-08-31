package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"sqlite/model"
)

//go:embed db/migrations/*.sql
var migrationFS embed.FS

type DB struct {
	*model.Queries
	db *sql.DB
}

func CreateAndMigrateDb(ctx context.Context, dsn string) (*DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot open db: %w", err)
	}
	if err := migrate(ctx, db); err != nil {
		// if we can't migrate, close the DB
		db.Close()
		return nil, fmt.Errorf("cannot migrate db: %w", err)
	}
	return &DB{
		Queries: model.New(db),
		db:      db,
	}, nil
}

func (db *DB) Transaction(ctx context.Context, run func(context.Context, *model.Queries) error) error {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	queries := db.Queries.WithTx(tx)
	err = run(ctx, queries)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) Close() error {
	return db.db.Close()
}

func migrate(ctx context.Context, db *sql.DB) error {

	// set up the connection
	if _, err := db.ExecContext(ctx, `
		PRAGMA busy_timeout = 5000;
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;
		PRAGMA foreign_keys = ON;
	`); err != nil {
		return fmt.Errorf("cannot set up pragmas: %w", err)
	}

	if _, err := db.ExecContext(ctx, `create table if not exists migrations (name text primary key);`); err != nil {
		return fmt.Errorf("cannot create migrations table: %w", err)
	}

	// get all saved migrations from DB and build a map
	rows, err := db.QueryContext(ctx, `select name from migrations`)
	if err != nil {
		return fmt.Errorf("cannot load migrations table: %w", err)
	}
	migrations := map[string]bool{}
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return fmt.Errorf("cannot scan migration row: %w", err)
		}

		migrations[name] = true
	}

	names, err := fs.Glob(migrationFS, "db/migrations/*.sql")
	if err != nil {
		return fmt.Errorf("cannot load migration files: %w", err)
	}
	sort.Strings(names)
	for _, name := range names {
		// run only migrations that aren't already saved in the DB
		if !migrations[name] {
			if err = migrateFile(db, name); err != nil {
				return fmt.Errorf("cannot migrate file %s: %w", name, err)
			}
		}
	}
	return nil
}

func migrateFile(db *sql.DB, name string) error {
	fmt.Printf("applying migration %s\n", name)
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Read and execute migration file.
	if buf, err := fs.ReadFile(migrationFS, name); err != nil {
		return err
	} else if _, err := tx.Exec(string(buf)); err != nil {
		return err
	}

	// Insert record into migrations to prevent re-running migration.
	if _, err := tx.Exec(`insert into migrations (name) values (?)`, name); err != nil {
		return err
	}

	return tx.Commit()
}
