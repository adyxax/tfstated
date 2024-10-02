package database

import (
	"encoding/json"
)

// Atomically check if there is an existing lock in place on the state. Returns
// true if it can be set, otherwise returns false and lock is set to the value
// of the existing lock
func (db *DB) SetLockOrGetExistingLock(name string, lock any) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var data []byte
	if err = tx.QueryRowContext(db.ctx, `SELECT lock FROM states WHERE name = ?;`, name).Scan(&data); err != nil {
		return false, err
	}
	if data != nil {
		err = json.Unmarshal(data, lock)
		return false, err
	}
	if data, err = json.Marshal(lock); err != nil {
		return false, err
	}
	_, err = tx.Exec(`UPDATE states SET lock = json(?) WHERE name = ?;`, data, name)
	if err != nil {
		return false, err
	}
	err = tx.Commit()
	return true, err
}

func (db *DB) Unlock(name, lock any) (bool, error) {
	data, err := json.Marshal(lock)
	if err != nil {
		return false, err
	}
	result, err := db.Exec(`UPDATE states SET lock = NULL WHERE name = ? and lock = json(?);`, name, data)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n == 1, err
}
