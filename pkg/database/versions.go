package database

import (
	"fmt"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/model"
)

func (db *DB) LoadVersionsByState(state *model.State) ([]model.Version, error) {
	rows, err := db.Query(
		`SELECT account_id, created, data, id, lock
           FROM versions
           WHERE state_id = ?
           ORDER BY id DESC;`, state.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to load versions from database: %w", err)
	}
	defer rows.Close()
	versions := make([]model.Version, 0)
	for rows.Next() {
		var version model.Version
		var created int64
		err = rows.Scan(&version.AccountId, &created, &version.Data, &version.Id, &version.Lock)
		if err != nil {
			return nil, fmt.Errorf("failed to load version from row: %w", err)
		}
		version.Created = time.Unix(created, 0)
		versions = append(versions, version)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to load versions from rows: %w", err)
	}
	return versions, nil
}
