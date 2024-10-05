package database

import (
	"database/sql"
	"fmt"
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
	err := db.QueryRow(`SELECT data FROM states WHERE name = ?;`, name).Scan(&encryptedData)
	if err != nil {
		return nil, err
	}
	return db.dataEncryptionKey.DecryptAES256(encryptedData)
}

// returns true in case of id mismatch
func (db *DB) SetState(name string, data []byte, id string) (bool, error) {
	encryptedData, err := db.dataEncryptionKey.EncryptAES256(data)
	if err != nil {
		return false, err
	}
	if id == "" {
		_, err = db.Exec(
			`INSERT INTO states(name, data) VALUES (:name, :data) ON CONFLICT DO UPDATE SET data = :data WHERE name = :name;`,
			sql.Named("data", encryptedData),
			sql.Named("name", name),
		)
		return false, err
	} else {
		result, err := db.Exec(`UPDATE states SET data = ? WHERE name = ? and lock->>'ID' = ?;`, encryptedData, name, id)
		if err != nil {
			return false, err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return false, err
		}
		if n != 1 {
			return true, fmt.Errorf("failed to update state, lock ID does not match")
		}
		return false, nil
	}
}
