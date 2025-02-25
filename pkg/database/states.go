package database

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"github.com/mattn/go-sqlite3"
	"go.n16f.net/uuid"
)

func (db *DB) CreateState(path string, accountId string, data []byte) (*model.Version, error) {
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
	)
	err := db.QueryRow(
		`SELECT created, lock, path, updated FROM states WHERE id = ?;`,
		stateId).Scan(&created, &state.Lock, &state.Path, &updated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load state id %s from database: %w", stateId, err)
	}
	state.Created = time.Unix(created, 0)
	state.Updated = time.Unix(updated, 0)
	return &state, nil
}

func (db *DB) LoadStates() ([]model.State, error) {
	rows, err := db.Query(
		`SELECT created, id, lock, path, updated FROM states;`)
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
		)
		err = rows.Scan(&created, &state.Id, &state.Lock, &state.Path, &updated)
		if err != nil {
			return nil, fmt.Errorf("failed to load state from row: %w", err)
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

// returns true in case of lock mismatch
func (db *DB) SetState(path string, accountId string, data []byte, lock string) (bool, error) {
	encryptedData, err := db.dataEncryptionKey.EncryptAES256(data)
	if err != nil {
		return false, fmt.Errorf("failed to encrypt state data: %w", err)
	}
	ret := false
	return ret, db.WithTransaction(func(tx *sql.Tx) error {
		var (
			stateId  string
			lockData []byte
		)
		if err = tx.QueryRowContext(db.ctx, `SELECT id, lock->>'ID' FROM states WHERE path = ?;`, path).Scan(&stateId, &lockData); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				var stateUUID uuid.UUID
				if err := stateUUID.Generate(uuid.V7); err != nil {
					return fmt.Errorf("failed to generate state id: %w", err)
				}
				_, err = tx.ExecContext(db.ctx, `INSERT INTO states(id, path) VALUES (?, ?)`, stateUUID, path)
				if err != nil {
					return fmt.Errorf("failed to insert new state: %w", err)
				}
				stateId = stateUUID.String()
			} else {
				return err
			}
		}

		if lock != "" && slices.Compare([]byte(lock), lockData) != 0 {
			err = fmt.Errorf("failed to update state: lock ID mismatch")
			ret = true
			return err
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
		_, err = tx.ExecContext(db.ctx,
			`DELETE FROM versions
           WHERE state_id = (SELECT id
                               FROM states
                               WHERE path = :path)
             AND id < (SELECT MIN(id)
                         FROM(SELECT versions.id
                                FROM versions
                                JOIN states ON states.id = versions.state_id
                                WHERE states.path = :path
                                ORDER BY versions.id DESC
                                LIMIT :limit));`,
			sql.Named("limit", db.versionsHistoryLimit),
			sql.Named("path", path),
		)
		return err
	})
}
