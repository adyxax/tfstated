package database

import (
	"database/sql"
	"encoding/json"
	"errors"
)

// Atomically check the lock status of a state and lock it if unlocked. Returns
// true if the function locked the state, otherwise returns false and the lock
// parameter is updated to the value of the existing lock
func (db *DB) SetLockOrGetExistingLock(path string, lock any) (bool, error) {
	ret := false
	return ret, db.WithTransaction(func(tx *sql.Tx) error {
		var lockData []byte
		if err := tx.QueryRowContext(db.ctx, `SELECT lock FROM states WHERE path = ?;`, path).Scan(&lockData); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if lockData, err = json.Marshal(lock); err != nil {
					return err
				}
				_, err = tx.ExecContext(db.ctx, `INSERT INTO states(path, lock) VALUES (?, json(?))`, path, lockData)
				ret = true
				return err
			} else {
				return err
			}
		}
		if lockData != nil {
			return json.Unmarshal(lockData, lock)
		}
		var err error
		if lockData, err = json.Marshal(lock); err != nil {
			return err
		}
		_, err = tx.ExecContext(db.ctx, `UPDATE states SET lock = json(?) WHERE path = ?;`, lockData, path)
		ret = true
		return err
	})
}

func (db *DB) Unlock(path, lock any) (bool, error) {
	data, err := json.Marshal(lock)
	if err != nil {
		return false, err
	}
	result, err := db.Exec(`UPDATE states SET lock = NULL WHERE path = ? and lock = json(?);`, path, data)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n == 1, err
}
