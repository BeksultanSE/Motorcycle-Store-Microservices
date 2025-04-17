package domain

import "errors"

var (
	ErrInvalidEmail    = errors.New("Invalid email")
	ErrInvalidPassword = errors.New("Invalid password")
	ErrUserNotFound    = errors.New("User not found")
	ErrUserExists      = errors.New("User already exists")
)
