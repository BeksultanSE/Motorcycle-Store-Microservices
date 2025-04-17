package usecase

import (
	"context"
	"github.com/BeksultanSE/Assignment1-user/internal/domain"
)

type AutoIncRepo interface {
	Next(ctx context.Context, coll string) (uint64, error)
}

type UserRepo interface {
	Create(ctx context.Context, user domain.User) error
	GetWithFilter(ctx context.Context, filter domain.UserFilter) (domain.User, error)
	Update(ctx context.Context, filter domain.UserFilter, update domain.UserUpdate) error
	Delete(ctx context.Context, filter domain.UserFilter) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) bool
}
