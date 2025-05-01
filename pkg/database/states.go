package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"github.com/mattn/go-sqlite3"
	"go.n16f.net/uuid"
)

func (db *DB) CreateState(path string, accountId uuid.UUID, data []byte) (*model.Version, error) {
	encryptedData, err := db.dataEncryptionKey.EncryptAES256(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt state data: %w", err)
	}
	var stateId uuid.UUID
	if err := stateId.Generate(uuid.V7); err != nil {
		return nil, fmt.Errorf("failed to generate state id: %w", err)
	}
	var versionId uuid.UUID
	if err := versionId.Generate(uuid.V7); err != nil {
		return nil, fmt.Errorf("failed to generate version id: %w", err)
	}
	version := &model.Version{
		AccountId: accountId,
		Id:        versionId,
		StateId:   stateId,
	}
	return version, db.WithTransaction(func(tx *sql.Tx) error {
		_, err := tx.ExecContext(db.ctx, `INSERT INTO states(id, path) VALUES (?, ?)`, stateId, path)
		if err != nil {
			var sqliteErr sqlite3.Error
			if errors.As(err, &sqliteErr) {
				if sqliteErr.Code == sqlite3.ErrNo(sqlite3.ErrConstraint) {
					version = nil
					return nil
				}
			}
			return fmt.Errorf("failed to insert new state: %w", err)
		}
		_, err = tx.ExecContext(db.ctx,
			`INSERT INTO versions(id, account_id, data, state_id)
               VALUES (:id, :accountID, :data, :stateID)`,
			sql.Named("accountID", accountId),
			sql.Named("data", encryptedData),
			sql.Named("id", versionId),
			sql.Named("stateID", stateId))
		if err != nil {
			return fmt.Errorf("failed to insert new state version: %w", err)
		}
		return nil
	})
}

// returns true in case of successful deletion
func (db *DB) DeleteState(path string) (bool, error) {
	result, err := db.Exec(`DELETE FROM states WHERE path = ?;`, path)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

func (db *DB) GetState(path string) ([]byte, error) {
	var encryptedData []byte
	err := db.QueryRow(
		`SELECT versions.data
           FROM versions
           JOIN states ON states.id = versions.state_id
           WHERE states.path = ?
           ORDER BY versions.id DESC
           LIMIT 1;`,
		path).Scan(&encryptedData)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []byte{}, nil
		}
		return nil, err
	}
	if encryptedData == nil {
		return []byte{}, nil
	}
	return db.dataEncryptionKey.DecryptAES256(encryptedData)
}

func (db *DB) LoadStateById(stateId uuid.UUID) (*model.State, error) {
	state := model.State{
		Id: stateId,
	}
	var (
		created int64
		updated int64
		lock    []byte
	)
	err := db.QueryRow(
		`SELECT created, json_extract(lock, '$'), path, updated
           FROM states
           WHERE id = ?;`,
		stateId).Scan(&created, &lock, &state.Path, &updated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load state id %s from database: %w", stateId, err)
	}
	if lock != nil {
		if err := json.Unmarshal(lock, &state.Lock); err != nil {
			return nil, fmt.Errorf("failed to unmarshal lock data: %w", err)
		}
	}
	state.Created = time.Unix(created, 0)
	state.Updated = time.Unix(updated, 0)
	return &state, nil
}

func (db *DB) LoadStatePaths() (map[string]string, error) {
	rows, err := db.Query(
		`SELECT id, path FROM states;`)
	if err != nil {
		return nil, fmt.Errorf("failed to load states from database: %w", err)
	}
	defer rows.Close()
	states := make(map[string]string)
	for rows.Next() {
		var (
			id   string
			path string
		)
		err = rows.Scan(&id, &path)
		if err != nil {
			return nil, fmt.Errorf("failed to load state from row: %w", err)
		}
		states[id] = path
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to load states from rows: %w", err)
	}
	return states, nil
}

func (db *DB) LoadStates() ([]model.State, error) {
	rows, err := db.Query(
		`SELECT created, id, json_extract(lock, '$'), path, updated FROM states;`)
	if err != nil {
		return nil, fmt.Errorf("failed to load states from database: %w", err)
	}
	defer rows.Close()
	states := make([]model.State, 0)
	for rows.Next() {
		var (
			state   model.State
			created int64
			updated int64
			lock    []byte
		)
		err = rows.Scan(&created, &state.Id, &lock, &state.Path, &updated)
		if err != nil {
			return nil, fmt.Errorf("failed to load state from row: %w", err)
		}
		if lock != nil {
			if err := json.Unmarshal(lock, &state.Lock); err != nil {
				return nil, fmt.Errorf("failed to unmarshal lock data: %w", err)
			}
		}
		state.Created = time.Unix(created, 0)
		state.Updated = time.Unix(updated, 0)
		states = append(states, state)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to load states from rows: %w", err)
	}
	return states, nil
}

// Returns (true, nil) on successful save
func (db *DB) SaveState(state *model.State) (bool, error) {
	lock, err := json.Marshal(state.Lock)
	if err != nil {
		return false, fmt.Errorf("failed to marshal lock data: %w", err)
	}
	_, err = db.Exec(
		`UPDATE states
           SET lock = jsonb(?),
               path = ?
           WHERE id = ?`,
		lock,
		state.Path,
		state.Id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code == sqlite3.ErrNo(sqlite3.ErrConstraint) {
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to update state id %s: %w", state.Id, err)
	}
	return true, nil
}

// returns true in case of lock mismatch
func (db *DB) SetState(path string, accountId uuid.UUID, data []byte, lockId string) (bool, error) {
	encryptedData, err := db.dataEncryptionKey.EncryptAES256(data)
	if err != nil {
		return false, fmt.Errorf("failed to encrypt state data: %w", err)
	}
	ret := false
	return ret, db.WithTransaction(func(tx *sql.Tx) error {
		var (
			stateId  string
			lockData *string
		)
		if err := tx.QueryRowContext(db.ctx, `SELECT id, lock->>'ID' FROM states WHERE path = ?;`, path).Scan(&stateId, &lockData); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				var stateUUID uuid.UUID
				if err := stateUUID.Generate(uuid.V7); err != nil {
					return fmt.Errorf("failed to generate state id: %w", err)
				}
				_, err := tx.ExecContext(db.ctx, `INSERT INTO states(id, path) VALUES (?, ?)`, stateUUID, path)
				if err != nil {
					return fmt.Errorf("failed to insert new state: %w", err)
				}
				stateId = stateUUID.String()
			} else {
				return fmt.Errorf("failed to select lock data from state: %w", err)
			}
		}

		if lockId != "" && (lockData == nil || lockId != *lockData) {
			ret = true
			return fmt.Errorf("failed to update state: lock ID mismatch")
		}
		var versionId uuid.UUID
		if err := versionId.Generate(uuid.V7); err != nil {
			return fmt.Errorf("failed to generate version id: %w", err)
		}
		_, err = tx.ExecContext(db.ctx,
			`INSERT INTO versions(id, account_id, state_id, data, lock)
               SELECT :versionId, :accountId, :stateId, :data, lock
                 FROM states
                 WHERE states.id = :stateId;`,
			sql.Named("accountId", accountId),
			sql.Named("data", encryptedData),
			sql.Named("stateId", stateId),
			sql.Named("versionId", versionId))
		if err != nil {
			return fmt.Errorf("failed to insert new state version: %w", err)
		}
		_, err = tx.ExecContext(db.ctx,
			`UPDATE states SET updated = ? WHERE id = ?;`,
			time.Now().UTC().Unix(),
			stateId)
		if err != nil {
			return fmt.Errorf("failed to touch updated for state: %w", err)
		}
		min := time.Now().Add(time.Duration(db.versionsHistoryMinimumDays) * -24 * time.Hour)
		_, err = tx.ExecContext(db.ctx,
			`DELETE FROM versions
               WHERE state_id = (SELECT id
                                   FROM states
                                   WHERE path = :path)
               AND id < (SELECT MIN(id)
                           FROM(SELECT versions.id
                                  FROM versions
                                  JOIN states ON states.id = versions.state_id
                                  WHERE states.path = :path AND versions.created < :min
                                  ORDER BY versions.id DESC
                                  LIMIT :limit));`,
			sql.Named("limit", db.versionsHistoryLimit),
			sql.Named("min", min),
			sql.Named("path", path),
		)
		return err
	})
}
