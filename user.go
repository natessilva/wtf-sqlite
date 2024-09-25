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
	return svc.db.Queries.GetUserById(ctx, UserFromFromContext(ctx).UserID)
}
