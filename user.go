package sqlite

import (
	"context"
	"sqlite/model"
)

type UserService struct {
	db *DB
}

func NewUserService(db *DB) *UserService {
	return &UserService{
		db: db,
	}
}

func (svc *UserService) Get(ctx context.Context) (model.User, error) {
	userID := UserFromFromContext(ctx)
	user, err := svc.db.Queries.GetUserById(ctx, userID)
	if err != nil {
		return model.User{}, err
	}
	return user, err
}
