package dao

import (
	"github.com/BeksultanSE/Assignment1-user/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

type User struct {
	ID             uint64    `bson:"_id"`
	Name           string    `bson:"name"`
	Email          string    `bson:"email"`
	HashedPassword string    `bson:"hashed_password"`
	CreatedAt      time.Time `bson:"createdAt"`
	UpdatedAt      time.Time `bson:"updatedAt"`
}

// FromUser converts user model to user dao for mongo
func FromUser(user domain.User) User {
	return User{
		ID:             user.ID,
		Name:           user.Name,
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}
}

// ToUser converts dao user to user model
func ToUser(user User) domain.User {
	return domain.User{
		ID:             user.ID,
		Name:           user.Name,
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}
}

// FromUserFilter constructs filtering query for mongo
func FromUserFilter(filter domain.UserFilter) bson.M {
	query := bson.M{}

	if filter.ID != nil {
		query["_id"] = filter.ID
	}
	if filter.Name != nil {
		query["name"] = filter.Name
	}
	if filter.Email != nil {
		query["email"] = filter.Email
	}

	return query
}

// FromUserUpdate constructs updating query for mongo
func FromUserUpdate(update domain.UserUpdate) bson.M {
	query := bson.M{}

	if update.Name != nil {
		query["name"] = update.Name
	}
	if update.Email != nil {
		query["email"] = update.Email
	}
	if update.HashedPassword != nil {
		query["hashed_password"] = update.HashedPassword
	}

	return bson.M{"$set": query}
}
