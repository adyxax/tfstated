package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/helpers"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"github.com/mattn/go-sqlite3"
	"go.n16f.net/uuid"
)

// Overriden by tests
var AdvertiseAdminPassword = func(password string) {
	slog.Info("Generated an initial admin password, please change it or delete the admin account after your first login", "password", password)
}

func (db *DB) CreateAccount(username string, isAdmin bool) (*model.Account, error) {
	var accountId uuid.UUID
	if err := accountId.Generate(uuid.V7); err != nil {
		return nil, fmt.Errorf("failed to generate account id: %w", err)
	}
	var passwordReset uuid.UUID
	if err := passwordReset.Generate(uuid.V4); err != nil {
		return nil, fmt.Errorf("failed to generate password reset uuid: %w", err)
	}
	_, err := db.Exec(
		`INSERT INTO accounts(id, username, is_Admin, settings, password_reset)
           VALUES (?, ?, ?, jsonb(?), ?);`,
		accountId,
		username,
		isAdmin,
		[]byte("{}"),
		passwordReset,
	)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code == sqlite3.ErrNo(sqlite3.ErrConstraint) {
				return nil, nil
			}
		}
		return nil, fmt.Errorf("failed to insert new account: %w", err)
	}
	return &model.Account{
		Id:            accountId,
		Username:      username,
		IsAdmin:       isAdmin,
		PasswordReset: &passwordReset,
	}, nil
}

func (db *DB) InitAdminAccount() error {
	return db.WithTransaction(func(tx *sql.Tx) error {
		var hasAdminAccount bool
		if err := tx.QueryRowContext(db.ctx, `SELECT EXISTS (SELECT 1 FROM accounts WHERE is_admin);`).Scan(&hasAdminAccount); err != nil {
			return fmt.Errorf("failed to select if there is an admin account in the database: %w", err)
		}
		if !hasAdminAccount {
			var accountId uuid.UUID
			if err := accountId.Generate(uuid.V7); err != nil {
				return fmt.Errorf("failed to generate account id: %w", err)
			}
			var password uuid.UUID
			if err := password.Generate(uuid.V4); err != nil {
				return fmt.Errorf("failed to generate initial admin password: %w", err)
			}
			salt := helpers.GenerateSalt()
			hash := helpers.HashPassword(password.String(), salt)
			if _, err := tx.ExecContext(db.ctx,
				`INSERT INTO accounts(id, username, salt, password_hash, is_admin, settings)
		       VALUES (:id, "admin", :salt, :hash, TRUE, jsonb(:settings))
		       ON CONFLICT DO UPDATE SET password_hash = :hash
		         WHERE username = "admin";`,
				sql.Named("id", accountId),
				sql.Named("hash", hash),
				sql.Named("salt", salt),
				sql.Named("settings", []byte("{}")),
			); err == nil {
				AdvertiseAdminPassword(password.String())
			} else {
				return fmt.Errorf("failed to set initial admin password: %w", err)
			}
		}
		return nil
	})
}

func (db *DB) LoadAccounts() ([]model.Account, error) {
	rows, err := db.Query(
		`SELECT id, username, salt, password_hash, is_admin, created, last_login,
                json_extract(settings, '$'), password_reset FROM accounts;`)
	if err != nil {
		return nil, fmt.Errorf("failed to load accounts from database: %w", err)
	}
	defer rows.Close()
	accounts := make([]model.Account, 0)
	for rows.Next() {
		var (
			account   model.Account
			created   int64
			lastLogin int64
			settings  []byte
		)
		err = rows.Scan(
			&account.Id,
			&account.Username,
			&account.Salt,
			&account.PasswordHash,
			&account.IsAdmin,
			&created,
			&lastLogin,
			&settings,
			&account.PasswordReset)
		if err != nil {
			return nil, fmt.Errorf("failed to load account from row: %w", err)
		}
		if err := json.Unmarshal(settings, &account.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal account settings: %w", err)
		}
		account.Created = time.Unix(created, 0)
		account.LastLogin = time.Unix(lastLogin, 0)
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to load accounts from rows: %w", err)
	}
	return accounts, nil
}

func (db *DB) LoadAccountUsernames() (map[string]string, error) {
	rows, err := db.Query(
		`SELECT id, username FROM accounts;`)
	if err != nil {
		return nil, fmt.Errorf("failed to load accounts from database: %w", err)
	}
	defer rows.Close()
	accounts := make(map[string]string)
	for rows.Next() {
		var (
			id       string
			username string
		)
		err = rows.Scan(&id, &username)
		if err != nil {
			return nil, fmt.Errorf("failed to load account from row: %w", err)
		}
		accounts[id] = username
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to load accounts from rows: %w", err)
	}
	return accounts, nil
}

func (db *DB) LoadAccountById(id *uuid.UUID) (*model.Account, error) {
	if id == nil {
		return nil, nil
	}
	account := model.Account{
		Id: *id,
	}
	var (
		created   int64
		lastLogin int64
		settings  []byte
	)
	err := db.QueryRow(
		`SELECT username, salt, password_hash, is_admin, created, last_login,
                json_extract(settings, '$'), password_reset
           FROM accounts
           WHERE id = ?;`,
		id,
	).Scan(&account.Username,
		&account.Salt,
		&account.PasswordHash,
		&account.IsAdmin,
		&created,
		&lastLogin,
		&settings,
		&account.PasswordReset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load account by id %s: %w", id, err)
	}
	if err := json.Unmarshal(settings, &account.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account settings: %w", err)
	}
	account.Created = time.Unix(created, 0)
	account.LastLogin = time.Unix(lastLogin, 0)
	return &account, nil
}

func (db *DB) LoadAccountByUsername(username string) (*model.Account, error) {
	account := model.Account{
		Username: username,
	}
	var (
		created   int64
		lastLogin int64
		settings  []byte
	)
	err := db.QueryRow(
		`SELECT id, salt, password_hash, is_admin, created, last_login,
                json_extract(settings, '$'), password_reset
           FROM accounts
           WHERE username = ?;`,
		username,
	).Scan(&account.Id,
		&account.Salt,
		&account.PasswordHash,
		&account.IsAdmin,
		&created,
		&lastLogin,
		&settings,
		&account.PasswordReset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load account by username %s: %w", username, err)
	}
	if err := json.Unmarshal(settings, &account.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account settings: %w", err)
	}
	account.Created = time.Unix(created, 0)
	account.LastLogin = time.Unix(lastLogin, 0)
	return &account, nil
}

func (db *DB) SaveAccount(account *model.Account) error {
	_, err := db.Exec(
		`UPDATE accounts
           SET username = ?,
               salt = ?,
               password_hash = ?,
               is_admin = ?,
               password_reset = ?
           WHERE id = ?`,
		account.Username,
		account.Salt,
		account.PasswordHash,
		account.IsAdmin,
		account.PasswordReset,
		account.Id)
	if err != nil {
		return fmt.Errorf("failed to update account id %s: %w", account.Id, err)
	}
	return nil
}

func (db *DB) SaveAccountSettings(account *model.Account, settings *model.Settings) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings for user accont %s: %w", account.Username, err)
	}
	_, err = db.Exec(`UPDATE accounts SET settings = ? WHERE id = ?`, data, account.Id)
	if err != nil {
		return fmt.Errorf("failed to update account settings for user account %s: %w", account.Username, err)
	}
	_, err = db.Exec(
		`UPDATE sessions
           SET data = jsonb_replace(data,
                                    '$.settings', jsonb(:data),
                                    '$.account.settings', jsonb(:data))
           WHERE data->'account'->>'id' = :id`,
		sql.Named("data", data),
		sql.Named("id", account.Id))
	if err != nil {
		return fmt.Errorf("failed to update account settings for user account %s: %w", account.Username, err)
	}
	return nil
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
