package usecase

import (
	"context"
	"github.com/BeksultanSE/Assignment1-user/internal/domain"
)

type UserUsecase struct {
	aiRepo   AutoIncRepo
	userRepo UserRepo
	pHasher  PasswordHasher
}

func NewUserUsecase(ai AutoIncRepo, userRepo UserRepo, pHasher PasswordHasher) UserUsecase {
	return UserUsecase{
		aiRepo:   ai,
		userRepo: userRepo,
		pHasher:  pHasher,
	}
}

func (uc UserUsecase) Register(ctx context.Context, req domain.User) (domain.User, error) {

	emailFilter := domain.UserFilter{
		Email: &req.Email,
	}
	if exists, _ := uc.userRepo.GetWithFilter(ctx, emailFilter); exists != (domain.User{}) {
		return domain.User{}, domain.ErrUserExists
	}

	id, err := uc.aiRepo.Next(ctx, domain.UserDB)
	if err != nil {
		return domain.User{}, err
	}
	req.ID = id

	req.HashedPassword, err = uc.pHasher.Hash(req.HashedPassword)
	if err != nil {
		return domain.User{}, err
	}

	err = uc.userRepo.Create(ctx, req)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		ID:   id,
		Name: req.Name,
	}, nil
}

func (uc UserUsecase) Authenticate(ctx context.Context, req domain.User) (domain.User, error) {
	emailFilter := domain.UserFilter{
		Email: &req.Email,
	}
	existingUser, err := uc.userRepo.GetWithFilter(ctx, emailFilter)
	if err != nil {
		return domain.User{}, err
	}

	if existingUser == (domain.User{}) {
		return domain.User{}, domain.ErrUserNotFound
	}

	isValid := uc.pHasher.Verify(existingUser.HashedPassword, req.HashedPassword)
	if !isValid {
		return domain.User{}, domain.ErrInvalidPassword
	}

	return domain.User{
		ID:   existingUser.ID,
		Name: existingUser.Name,
	}, nil
}

func (uc UserUsecase) Get(ctx context.Context, filter domain.UserFilter) (domain.User, error) {
	user, err := uc.userRepo.GetWithFilter(ctx, filter)
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}
