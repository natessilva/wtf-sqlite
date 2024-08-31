package sqlite

import (
	"context"
	"sqlite/model"
)

type contextKey struct{}

var key contextKey

type AuthenticatedUser struct {
	UserID int64
}

func ContextWithUser(ctx context.Context, i AuthenticatedUser) context.Context {
	return context.WithValue(ctx, &key, i)
}

func UserFromContext(ctx context.Context) AuthenticatedUser {
	return ctx.Value(&key).(AuthenticatedUser)
}

type UserService struct {
	db *DB
}

func NewUserService(db *DB) *UserService {
	return &UserService{
		db: db,
	}
}

func (svc *UserService) Get(ctx context.Context) (model.User, error) {
	userID := UserFromContext(ctx).UserID
	user, err := svc.db.Queries.GetUserById(ctx, userID)
	if err != nil {
		return model.User{}, err
	}
	return user, err
}
