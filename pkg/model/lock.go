package model

import "time"

type Lock struct {
	Created   time.Time `json:"Created"`
	Id        string    `json:"ID"`
	Info      string    `json:"Info"`
	Operation string    `json:"Operation"`
	Path      string    `json:"Path"`
	Version   string    `json:"Version"`
	Who       string    `json:"Who"`
}
