package domain

import "time"

type Product struct {
	ID        uint64
	Name      string
	Category  string
	Price     float64
	Stock     uint64
	CreatedAt time.Time
	UpdatedAt time.Time
}
