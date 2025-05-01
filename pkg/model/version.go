package model

import (
	"encoding/json"
	"time"

	"go.n16f.net/uuid"
)

type Version struct {
	AccountId uuid.UUID
	Created   time.Time
	Data      json.RawMessage
	Id        uuid.UUID
	Lock      *Lock
	StateId   uuid.UUID
}
