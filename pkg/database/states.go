package database

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
)

// returns true in case of successful deletion
func (db *DB) DeleteState(name string) (bool, error) {
	result, err := db.Exec(`DELETE FROM states WHERE name = ?;`, name)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

func (db *DB) GetState(name string) ([]byte, error) {
	var encryptedData []byte
	err := db.QueryRow(
		`SELECT versions.data
           FROM versions
           JOIN states ON states.id = versions.state_id
           WHERE states.name = ?
           ORDER BY versions.id DESC
           LIMIT 1;`,
		name).Scan(&encryptedData)
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

// returns true in case of id mismatch
func (db *DB) SetState(name string, data []byte, lockID string) (bool, error) {
	encryptedData, err := db.dataEncryptionKey.EncryptAES256(data)
	if err != nil {
		return false, err
	}
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var (
		stateID  int64
		lockData []byte
	)
	if err = tx.QueryRowContext(db.ctx, `SELECT id, lock->>'ID' FROM states WHERE name = ?;`, name).Scan(&stateID, &lockData); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			var result sql.Result
			result, err = tx.ExecContext(db.ctx, `INSERT INTO states(name) VALUES (?)`, name)
			if err != nil {
				return false, err
			}
			stateID, err = result.LastInsertId()
			if err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	}

	if lockID != "" && slices.Compare([]byte(lockID), lockData) != 0 {
		err = fmt.Errorf("failed to update state, lock ID does not match")
		return true, err
	}
	_, err = tx.ExecContext(db.ctx,
		`INSERT INTO versions(state_id, data, lock)
           SELECT :stateID, :data, lock
             FROM states
             WHERE states.id = :stateID;`,
		sql.Named("stateID", stateID),
		sql.Named("data", encryptedData))
	if err != nil {
		return false, err
	}
	_, err = tx.ExecContext(db.ctx,
		`DELETE FROM versions
           WHERE state_id = (SELECT id
                               FROM states
                               WHERE name = :name)
             AND id < (SELECT MIN(id)
                         FROM(SELECT versions.id
                                FROM versions
                                JOIN states ON states.id = versions.state_id
                                WHERE states.name = :name
                                ORDER BY versions.id DESC
                                LIMIT :limit));`,
		sql.Named("limit", db.versionsHistoryLimit),
		sql.Named("name", name),
	)
	if err != nil {
		return false, err
	}
	err = tx.Commit()
	return false, err
}
