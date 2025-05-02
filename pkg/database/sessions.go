package database

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/helpers"
	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"git.adyxax.org/adyxax/tfstated/pkg/scrypto"
)

func (db *DB) CreateSession(sessionData *model.SessionData) (string, *model.Session, error) {
	sessionBytes := scrypto.RandomBytes(32)
	sessionId := base64.RawURLEncoding.EncodeToString(sessionBytes[:])
	sessionHash := helpers.HashSessionId(sessionBytes, db.sessionsSalt.Bytes())
	if sessionData == nil {
		var err error
		sessionData, err = model.NewSessionData(nil, nil)
		if err != nil {
			return "", nil, fmt.Errorf("failed to generate new session data: %w", err)
		}
	}
	data, err := json.Marshal(sessionData)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal session data: %w", err)
	}
	if _, err := db.Exec(
		`INSERT INTO sessions(id, data)
		   VALUES (?, jsonb(?));`,
		sessionHash,
		data,
	); err != nil {
		return "", nil, fmt.Errorf("failed insert new session in database: %w", err)
	}
	return sessionId, &model.Session{
		Id:   sessionHash,
		Data: sessionData,
	}, nil
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
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (db *DB) DeleteSessions(account *model.Account) error {
	_, err := db.Exec(
		`DELETE FROM sessions WHERE data->'account'->>'id' = ?`,
		account.Id)
	if err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
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
		data    []byte
	)
	err = db.QueryRow(
		`SELECT created,
                updated,
                json_extract(data, '$')
           FROM sessions
           WHERE id = ?;`,
		sessionHash,
	).Scan(
		&created,
		&updated,
		&data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load session by id %s: %w", id, err)
	}
	if err := json.Unmarshal(data, &session.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}
	session.Created = time.Unix(created, 0)
	session.Updated = time.Unix(updated, 0)
	return &session, nil
}

func (db *DB) MigrateSession(session *model.Session, account *model.Account) (string, *model.Session, error) {
	if err := db.DeleteSession(session); err != nil {
		return "", nil, fmt.Errorf("failed to delete session: %w", err)
	}
	sessionData, err := model.NewSessionData(account, session.Data.Settings)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate new session data: %w", err)
	}
	sessionId, session, err := db.CreateSession(sessionData)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create session: %w", err)
	}
	return sessionId, session, nil
}
