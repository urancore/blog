package models

import "time"

type User struct {
	ID        int64
	Username  string
	Password  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
