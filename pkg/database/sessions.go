package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/helpers"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"git.adyxax.org/adyxax/tfstated/pkg/scrypto"
	"go.n16f.net/uuid"
)

func (db *DB) CreateSession(account *model.Account, settingsData []byte) (string, error) {
	sessionBytes := scrypto.RandomBytes(32)
	sessionId := base64.RawURLEncoding.EncodeToString(sessionBytes[:])
	sessionHash := helpers.HashSessionId(sessionBytes, db.sessionsSalt.Bytes())
	var accountId *uuid.UUID = nil
	var settings = []byte("{}")
	if account != nil {
		accountId = &account.Id
		settings = account.Settings
	} else if settingsData != nil {
		settings = settingsData
	}
	if _, err := db.Exec(
		`INSERT INTO sessions(id, account_id, settings)
		   VALUES (?, ?, ?);`,
		sessionHash,
		accountId,
		settings,
	); err != nil {
		return "", fmt.Errorf("failed insert new session in database: %w", err)
	}
	return sessionId, nil
}

func (db *DB) DeleteExpiredSessions() error {
	expires := time.Now().Add(-12 * time.Hour)
	_, err := db.Exec(`DELETE FROM sessions WHERE created < ?`, expires.Unix())
	if err != nil {
		return fmt.Errorf("failed to delete expired session: %w", err)
	}
	return nil
}

func (db *DB) DeleteSession(session *model.Session) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE id = ?`, session.Id)
	if err != nil {
		return fmt.Errorf("failed to delete session %s: %w", session.Id, err)
	}
	return nil
}

func (db *DB) LoadSessionById(id string) (*model.Session, error) {
	sessionBytes, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 session id: %w", err)
	}
	sessionHash := helpers.HashSessionId(sessionBytes, db.sessionsSalt.Bytes())
	session := model.Session{
		Id: sessionHash,
	}
	var (
		created int64
		updated int64
	)
	err = db.QueryRow(
		`SELECT account_id,
                created,
                updated,
                settings
           FROM sessions
           WHERE id = ?;`,
		sessionHash,
	).Scan(&session.AccountId,
		&created,
		&updated,
		&session.Settings,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load session by id %s: %w", id, err)
	}
	session.Created = time.Unix(created, 0)
	session.Updated = time.Unix(updated, 0)
	return &session, nil
}

func (db *DB) MigrateSession(session *model.Session, account *model.Account) (string, error) {
	if err := db.DeleteSession(session); err != nil {
		return "", fmt.Errorf("failed to delete session: %w", err)
	}
	sessionId, err := db.CreateSession(account, session.Settings)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	return sessionId, nil
}
