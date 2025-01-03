package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/helpers"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"go.n16f.net/uuid"
)

// Overriden by tests
var AdvertiseAdminPassword = func(password string) {
	slog.Info("Generated an initial admin password, please change it or delete the admin account after your first login", "password", password)
}

func (db *DB) InitAdminAccount() error {
	return db.WithTransaction(func(tx *sql.Tx) error {
		var hasAdminAccount bool
		if err := tx.QueryRowContext(db.ctx, `SELECT EXISTS (SELECT 1 FROM accounts WHERE is_admin);`).Scan(&hasAdminAccount); err != nil {
			return fmt.Errorf("failed to select if there is an admin account in the database: %w", err)
		}
		if !hasAdminAccount {
			var password uuid.UUID
			if err := password.Generate(uuid.V4); err != nil {
				return fmt.Errorf("failed to generate initial admin password: %w", err)
			}
			salt := helpers.GenerateSalt()
			hash := helpers.HashPassword(password.String(), salt)
			if _, err := tx.ExecContext(db.ctx,
				`INSERT INTO accounts(username, salt, password_hash, is_admin)
		       VALUES ("admin", :salt, :hash, TRUE)
		       ON CONFLICT DO UPDATE SET password_hash = :hash
		         WHERE username = "admin";`,
				sql.Named("salt", salt),
				sql.Named("hash", hash),
			); err == nil {
				AdvertiseAdminPassword(password.String())
			} else {
				return fmt.Errorf("failed to set initial admin password: %w", err)
			}
		}
		return nil
	})
}

func (db *DB) LoadAccountByUsername(username string) (*model.Account, error) {
	account := model.Account{
		Username: username,
	}
	var (
		created   int64
		lastLogin int64
	)
	err := db.QueryRow(
		`SELECT id, salt, password_hash, is_admin, created, last_login, settings
           FROM accounts
           WHERE username = ?;`,
		username,
	).Scan(&account.Id,
		&account.Salt,
		&account.PasswordHash,
		&account.IsAdmin,
		&created,
		&lastLogin,
		&account.Settings,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	account.Created = time.Unix(created, 0)
	account.LastLogin = time.Unix(lastLogin, 0)
	return &account, nil
}

func (db *DB) TouchAccount(account *model.Account) error {
	now := time.Now().UTC()
	_, err := db.Exec(`UPDATE accounts SET last_login = ? WHERE id = ?`, now.Unix(), account.Id)
	if err != nil {
		return fmt.Errorf("failed to update last_login for user %s: %w", account.Username, err)
	}
	account.LastLogin = now
	return nil
}
