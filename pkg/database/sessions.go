package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"go.n16f.net/uuid"
)

func (db *DB) CreateSession(account *model.Account) (string, error) {
	var sessionId uuid.UUID
	if err := sessionId.Generate(uuid.V4); err != nil {
		return "", fmt.Errorf("failed to generate session id: %w", err)
	}
	if _, err := db.Exec(
		`INSERT INTO sessions(id, account_id, data)
		   VALUES (?, ?, ?);`,
		sessionId.String(),
		account.Id,
		"",
	); err != nil {
		return "", fmt.Errorf("failed insert new session in database: %w", err)
	}
	return sessionId.String(), nil
}

func (db *DB) DeleteExpiredSessions() error {
	_, err := db.Exec(`DELETE FROM sessions WHERE created < ?`, time.Now().Unix())
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
	session := model.Session{
		Id: id,
	}
	var (
		created int64
		updated int64
	)
	err := db.QueryRow(
		`SELECT account_id,
                created,
                updated,
                data
           FROM sessions
           WHERE id = ?;`,
		id,
	).Scan(&session.AccountId,
		&created,
		&updated,
		&session.Data,
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

func (db *DB) TouchSession(sessionId string) error {
	now := time.Now().UTC()
	_, err := db.Exec(`UPDATE sessions SET updated = ? WHERE id = ?`, now.Unix(), sessionId)
	if err != nil {
		return fmt.Errorf("failed to touch updated for session %s: %w", sessionId, err)
	}
	return nil
}
