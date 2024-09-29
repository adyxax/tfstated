package database

import (
	"embed"
	"io/fs"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/*.sql
var schemaFiles embed.FS

func (db *DB) migrate() error {
	statements := make([]string, 0)
	err := fs.WalkDir(schemaFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() || err != nil {
			return err
		}
		var stmts []byte
		if stmts, err = schemaFiles.ReadFile(path); err != nil {
			return err
		} else {
			statements = append(statements, string(stmts))
		}
		return nil
	})
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var version int
	if err = tx.QueryRowContext(db.ctx, `SELECT version FROM schema_version;`).Scan(&version); err != nil {
		if err.Error() == "no such table: schema_version" {
			version = 0
		} else {
			return err
		}
	}

	for version < len(statements) {
		if _, err = tx.ExecContext(db.ctx, statements[version]); err != nil {
			return err
		}
		version++
	}
	if _, err = tx.ExecContext(db.ctx, `DELETE FROM schema_version; INSERT INTO schema_version (version) VALUES (?);`, version); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
