package database

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"

	"git.adyxax.org/adyxax/tfstated/pkg/scrypto"
)

func initDB(ctx context.Context, url string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = db.Close()
		}
	}()
	if _, err = db.ExecContext(ctx, "PRAGMA busy_timeout = 5000"); err != nil {
		return nil, err
	}

	return db, nil
}

type DB struct {
	ctx                  context.Context
	dataEncryptionKey    scrypto.AES256Key
	readDB               *sql.DB
	versionsHistoryLimit int
	writeDB              *sql.DB
}

func NewDB(ctx context.Context, url string) (*DB, error) {
	readDB, err := initDB(ctx, url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = readDB.Close()
		}
	}()
	readDB.SetMaxOpenConns(max(4, runtime.NumCPU()))

	writeDB, err := initDB(ctx, url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = writeDB.Close()
		}
	}()
	writeDB.SetMaxOpenConns(1)

	db := DB{
		ctx:                  ctx,
		readDB:               readDB,
		versionsHistoryLimit: 64,
		writeDB:              writeDB,
	}
	if _, err = db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}
	if _, err = db.Exec("PRAGMA cache_size = 10000000"); err != nil {
		return nil, err
	}
	if _, err = db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, err
	}
	if _, err = db.Exec("PRAGMA synchronous = NORMAL"); err != nil {
		return nil, err
	}
	if err = db.migrate(); err != nil {
		return nil, err
	}

	return &db, nil
}

func (db *DB) Close() error {
	if err := db.readDB.Close(); err != nil {
		_ = db.writeDB.Close()
	}
	return db.writeDB.Close()
}

func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	return db.writeDB.ExecContext(db.ctx, query, args...)
}

func (db *DB) QueryRow(query string, args ...any) *sql.Row {
	return db.readDB.QueryRowContext(db.ctx, query, args...)
}

func (db *DB) SetDataEncryptionKey(s string) error {
	return db.dataEncryptionKey.FromBase64(s)
}

func (db *DB) SetVersionsHistoryLimit(n int) {
	db.versionsHistoryLimit = n
}

func (db *DB) WithTransaction(f func(tx *sql.Tx) error) error {
	tx, err := db.writeDB.Begin()
	if err != nil {
		return err
	}
	err = f(tx)
	if err == nil {
		err = tx.Commit()
	}
	if err != nil {
		if err2 := tx.Rollback(); err2 != nil {
			panic(fmt.Sprintf("failed to rollback transaction: %+v. Reason for rollback: %+v", err2, err))
		}
	}
	return err
}
