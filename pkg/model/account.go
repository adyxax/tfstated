package model

import "time"

type AccountContextKey struct{}

type Account struct {
	Id        int
	Username  string
	Password  string
	IsAdmin   bool
	Created   time.Time
	LastLogin time.Time
	Settings  any
}
