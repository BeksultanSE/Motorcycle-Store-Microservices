package mongo

import (
	"context"
	"github.com/BeksultanSE/Assignment1-user/internal/adapter/mongo/dao"
	"github.com/BeksultanSE/Assignment1-user/internal/domain"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepo struct {
	conn       *mongo.Database
	collection string
}

func NewUserRepo(conn *mongo.Database) *UserRepo {
	return &UserRepo{
		conn:       conn,
		collection: CollectionUsers,
	}
}

func (u *UserRepo) Create(ctx context.Context, user domain.User) error {
	newUser := dao.FromUser(user)
	_, err := u.conn.Collection(u.collection).InsertOne(ctx, newUser)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserRepo) GetWithFilter(ctx context.Context, filter domain.UserFilter) (domain.User, error) {
	var userDao dao.User
	err := u.conn.Collection(u.collection).FindOne(
		ctx,
		dao.FromUserFilter(filter),
	).Decode(&userDao)
	if err != nil {
		return domain.User{}, err
	}
	return dao.ToUser(userDao), nil
}

func (u *UserRepo) Update(ctx context.Context, filter domain.UserFilter, update domain.UserUpdate) error {
	panic("implement me")
}
func (u *UserRepo) Delete(ctx context.Context, filter domain.UserFilter) error {
	panic("implement me")
}
