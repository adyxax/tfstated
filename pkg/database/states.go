package database

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

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

func (db *DB) LoadStateById(stateId int) (*model.State, error) {
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
		return nil, fmt.Errorf("failed to load state id %d from database: %w", stateId, err)
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
		var state model.State
		var (
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

// returns true in case of id mismatch
func (db *DB) SetState(path string, accountID string, data []byte, lockID string) (bool, error) {
	encryptedData, err := db.dataEncryptionKey.EncryptAES256(data)
	if err != nil {
		return false, fmt.Errorf("failed to encrypt state data: %w", err)
	}
	ret := false
	return ret, db.WithTransaction(func(tx *sql.Tx) error {
		var (
			stateID  int64
			lockData []byte
		)
		if err = tx.QueryRowContext(db.ctx, `SELECT id, lock->>'ID' FROM states WHERE path = ?;`, path).Scan(&stateID, &lockData); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				var result sql.Result
				result, err = tx.ExecContext(db.ctx, `INSERT INTO states(path) VALUES (?)`, path)
				if err != nil {
					return fmt.Errorf("failed to insert new state: %w", err)
				}
				stateID, err = result.LastInsertId()
				if err != nil {
					return fmt.Errorf("failed to get last insert id for new state: %w", err)
				}
			} else {
				return err
			}
		}

		if lockID != "" && slices.Compare([]byte(lockID), lockData) != 0 {
			err = fmt.Errorf("failed to update state, lock ID does not match")
			ret = true
			return err
		}
		_, err = tx.ExecContext(db.ctx,
			`INSERT INTO versions(account_id, state_id, data, lock)
           SELECT :accountID, :stateID, :data, lock
             FROM states
             WHERE states.id = :stateID;`,
			sql.Named("accountID", accountID),
			sql.Named("stateID", stateID),
			sql.Named("data", encryptedData))
		if err != nil {
			return fmt.Errorf("failed to insert new state version: %w", err)
		}
		_, err = tx.ExecContext(db.ctx,
			`UPDATE states SET updated = ? WHERE id = ?;`,
			time.Now().UTC().Unix(),
			stateID)
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
