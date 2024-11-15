package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"go.n16f.net/uuid"
)

func (db *DB) LoadAccountByUsername(username string) (*model.Account, error) {
	account := model.Account{
		Username: username,
	}
	var (
		encryptedPassword []byte
		created           int64
		lastLogin         int64
	)
	err := db.QueryRow(
		`SELECT id, password, is_admin, created, last_login, settings
           FROM accounts
           WHERE username = ?;`,
		username,
	).Scan(&account.Id,
		&encryptedPassword,
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
	password, err := db.dataEncryptionKey.DecryptAES256(encryptedPassword)
	if err != nil {
		return nil, err
	}
	account.Password = string(password)
	account.Created = time.Unix(created, 0)
	account.LastLogin = time.Unix(lastLogin, 0)
	return &account, nil
}

func (db *DB) InitAdminAccount() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var hasAdminAccount bool
	if err = tx.QueryRowContext(db.ctx, `SELECT EXISTS (SELECT 1 FROM accounts WHERE is_admin);`).Scan(&hasAdminAccount); err != nil {
		return fmt.Errorf("failed to select if there is an admin account in the database: %w", err)
	}
	if hasAdminAccount {
		tx.Rollback()
	} else {
		var password uuid.UUID
		if err = password.Generate(uuid.V4); err != nil {
			return fmt.Errorf("failed to generate initial admin password: %w", err)
		}
		var encryptedPassword []byte
		encryptedPassword, err = db.dataEncryptionKey.EncryptAES256([]byte(password.String()))
		if err != nil {
			return fmt.Errorf("failed to encrypt initial admin password: %w", err)
		}
		if _, err = tx.ExecContext(db.ctx,
			`INSERT INTO accounts(username, password, is_admin)
		       VALUES ("admin", :password, TRUE)
		       ON CONFLICT DO UPDATE SET password = :password
		         WHERE username = "admin";`,
			sql.Named("password", encryptedPassword),
		); err != nil {
			return fmt.Errorf("failed to set initial admin password: %w", err)
		}
		err = tx.Commit()
		if err == nil {
			slog.Info("Generated an initial admin password, please change it or delete the admin account after your first login", "password", password.String())
		}
	}
	return err
}
