package database

import (
	"database/sql"
)

func (db *DB) DeleteState(name string) error {
	_, err := db.Exec(`DELETE FROM states WHERE name = ?;`, name)
	return err
}

func (db *DB) GetState(name string) ([]byte, error) {
	var encryptedData []byte
	err := db.QueryRow(`SELECT data FROM states WHERE name = ?;`, name).Scan(&encryptedData)
	if err != nil {
		return nil, err
	}
	return db.dataEncryptionKey.DecryptAES256(encryptedData)
}

func (db *DB) SetState(name string, data []byte) error {
	encryptedData, err := db.dataEncryptionKey.EncryptAES256(data)
	if err != nil {
		return err
	}
	_, err = db.Exec(
		`INSERT INTO states(name, data) VALUES (:name, :data) ON CONFLICT DO UPDATE SET data = :data WHERE name = :name;`,
		sql.Named("data", encryptedData),
		sql.Named("name", name),
	)
	if err != nil {
		return err
	}
	return nil
}
