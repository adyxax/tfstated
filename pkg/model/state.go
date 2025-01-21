package model

import (
	"time"
)

type State struct {
	Created time.Time
	Id      int
	Lock    *string
	Path    string
	Updated time.Time
}
