package model

import (
	"time"
)

type Version struct {
	AccountId int
	Created   time.Time
	Data      []byte
	Id        int
	Lock      *string
	StateId   int
}
