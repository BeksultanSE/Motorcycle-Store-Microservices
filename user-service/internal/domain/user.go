package domain

import "time"

type User struct {
	ID             uint64
	Name           string
	Email          string
	HashedPassword string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type UserFilter struct {
	ID    *uint64
	Name  *string
	Email *string
}

type UserUpdate struct {
	Name           *string
	Email          *string
	HashedPassword *string
	UpdatedAt      *time.Time
}
