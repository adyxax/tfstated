package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/model"
	"go.n16f.net/uuid"
)

func (db *DB) LoadVersionById(id uuid.UUID) (*model.Version, error) {
	version := model.Version{
		Id: id,
	}
	var (
		created       int64
		encryptedData []byte
		lock          []byte
	)
	err := db.QueryRow(
		`SELECT account_id, state_id, data, json_extract(lock, '$'), created
           FROM versions WHERE id = ?;`,
		id).Scan(
		&version.AccountId,
		&version.StateId,
		&encryptedData,
		&lock,
		&created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load version id %s from database: %w", id, err)
	}
	if lock != nil {
		if err := json.Unmarshal(lock, &version.Lock); err != nil {
			return nil, fmt.Errorf("failed to unmarshal lock data: %w", err)
		}
	}
	version.Created = time.Unix(created, 0)
	version.Data, err = db.dataEncryptionKey.DecryptAES256(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt version %s data: %w", id, err)
	}
	return &version, nil
}

func (db *DB) LoadVersionsByState(state *model.State) ([]model.Version, error) {
	rows, err := db.Query(
		`SELECT account_id, created, data, id, json_extract(lock, '$')
           FROM versions
           WHERE state_id = ?
           ORDER BY id DESC;`, state.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to load versions from database: %w", err)
	}
	defer rows.Close()
	versions := make([]model.Version, 0)
	for rows.Next() {
		version := model.Version{StateId: state.Id}
		var created int64
		var lock []byte
		err = rows.Scan(&version.AccountId, &created, &version.Data, &version.Id, &lock)
		if err != nil {
			return nil, fmt.Errorf("failed to load version from row: %w", err)
		}
		if lock != nil {
			if err := json.Unmarshal(lock, &version.Lock); err != nil {
				return nil, fmt.Errorf("failed to unmarshal lock data: %w", err)
			}
		}
		version.Created = time.Unix(created, 0)
		versions = append(versions, version)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to load versions from rows: %w", err)
	}
	return versions, nil
}

func (db *DB) LoadVersionsByAccount(account *model.Account) ([]model.Version, error) {
	rows, err := db.Query(
		`SELECT created, data, id, json_extract(lock, '$'), state_id
           FROM versions
           WHERE account_id = ?
           ORDER BY id DESC;`, account.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to load versions from database: %w", err)
	}
	defer rows.Close()
	versions := make([]model.Version, 0)
	for rows.Next() {
		version := model.Version{AccountId: account.Id}
		var created int64
		var lock []byte
		err = rows.Scan(&created, &version.Data, &version.Id, &lock, &version.StateId)
		if err != nil {
			return nil, fmt.Errorf("failed to load version from row: %w", err)
		}
		if lock != nil {
			if err := json.Unmarshal(lock, &version.Lock); err != nil {
				return nil, fmt.Errorf("failed to unmarshal lock data: %w", err)
			}
		}
		version.Created = time.Unix(created, 0)
		versions = append(versions, version)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to load versions from rows: %w", err)
	}
	return versions, nil
}
