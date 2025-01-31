package model

import (
	"encoding/json"
	"time"
)

type Version struct {
	AccountId string
	Created   time.Time
	Data      json.RawMessage
	Id        int
	Lock      *string
	StateId   int
}
