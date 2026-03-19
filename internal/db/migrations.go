package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ApplyMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".sql") {
			files = append(files, name)
		}
	}
	sort.Strings(files)

	for _, name := range files {
		ver, err := parseVersion(name)
		if err != nil {
			return err
		}

		applied, err := isApplied(ctx, pool, ver)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		sqlBytes, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, string(sqlBytes))
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("migration %s failed: %w", name, err)
		}
		_, err = tx.Exec(ctx, "INSERT INTO schema_migrations(version) VALUES($1)", ver)
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("migration %s record failed: %w", name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}

func parseVersion(filename string) (int64, error) {
	// ожидаем что-то типа 001_init.sql
	base := strings.Split(filename, "_")[0]
	var v int64
	for _, ch := range base {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("bad migration filename: %s", filename)
		}
		v = v*10 + int64(ch-'0')
	}
	if v == 0 {
		return 0, fmt.Errorf("bad migration version: %s", filename)
	}
	return v, nil
}

func isApplied(ctx context.Context, pool *pgxpool.Pool, version int64) (bool, error) {
	// таблица может ещё не существовать
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_name='schema_migrations'
		)`).Scan(&exists)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	var ok bool

	err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)", version).Scan(&ok)
	return ok, err
}
