package models

import "time"

type APIPGS struct {
	Id          int
	KeyName     string
	CloudAccess string
	Email       string
	CreatedAt   time.Time
	LastLogin   time.Time
}
