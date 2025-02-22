package model

import (
	"time"

	"go.n16f.net/uuid"
)

type State struct {
	Created time.Time
	Id      uuid.UUID
	Lock    *string
	Path    string
	Updated time.Time
}
