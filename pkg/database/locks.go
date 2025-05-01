package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"go.n16f.net/uuid"
)

// Atomically check the lock status of a state and lock it if unlocked. Returns
// true if the function locked the state, otherwise returns false and the lock
// parameter is updated to the value of the existing lock
func (db *DB) SetLockOrGetExistingLock(path string, lock any) (bool, error) {
	ret := false
	return ret, db.WithTransaction(func(tx *sql.Tx) error {
		var lockData []byte
		err := tx.QueryRowContext(db.ctx,
			`SELECT json_extract(lock, '$')
               FROM states
               WHERE path = ?;`,
			path).Scan(&lockData)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if lockData, err = json.Marshal(lock); err != nil {
					return fmt.Errorf("failed to marshall lock data: %w", err)
				}
				var stateId uuid.UUID
				if err := stateId.Generate(uuid.V7); err != nil {
					return fmt.Errorf("failed to generate state id: %w", err)
				}
				_, err := tx.ExecContext(db.ctx,
					`INSERT INTO states(id, path, lock)
                       VALUES (?, ?, jsonb(?))`,
					stateId, path, lockData)
				if err != nil {
					return fmt.Errorf("failed to create new state: %w", err)
				}
				ret = true
				return nil
			}
			return fmt.Errorf("failed to select lock data from state: %w", err)
		}
		if lockData != nil {
			if err := json.Unmarshal(lockData, lock); err != nil {
				return fmt.Errorf("failed to unmarshal lock data: %w", err)
			}
			return nil
		}
		if lockData, err = json.Marshal(lock); err != nil {
			return fmt.Errorf("failed to marshal lock data: %w", err)
		}
		_, err = tx.ExecContext(db.ctx,
			`UPDATE states
               SET lock = jsonb(?)
               WHERE path = ?;`,
			lockData, path)
		if err != nil {
			return fmt.Errorf("failed to set lock data: %w", err)
		}
		ret = true
		return nil
	})
}

func (db *DB) Unlock(path string, lock any) (bool, error) {
	data, err := json.Marshal(lock)
	if err != nil {
		return false, fmt.Errorf("failed to marshal lock data: %w", err)
	}
	result, err := db.Exec(
		`UPDATE states
           SET lock = NULL
           WHERE path = ? and lock = jsonb(?);`,
		path, data)
	if err != nil {
		return false, fmt.Errorf("failed to update state: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get affected rows: %w", err)
	}
	return n == 1, nil
}
}
