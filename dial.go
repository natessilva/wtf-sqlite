package sqlite

import (
	"context"
	"sqlite/model"
)

type DialService struct {
	db *DB
}

func NewDialService(db *DB) *DialService {
	return &DialService{
		db: db,
	}
}

func (svc *DialService) Create(ctx context.Context, name string) (int64, error) {
	return svc.db.CreateDial(ctx, model.CreateDialParams{
		UserID: UserFromFromContext(ctx),
		Name:   name,
	})
}

func (svc *DialService) List(ctx context.Context) ([]model.Dial, error) {
	return svc.db.ListDials(ctx, UserFromFromContext(ctx))
}

func (svc *DialService) Get(ctx context.Context, id int64) (model.Dial, error) {
	return svc.db.GetDial(ctx, model.GetDialParams{
		UserID: UserFromFromContext(ctx),
		ID:     id,
	})
}

type UpdateDial struct {
	ID   int64
	Name string
}

func (svc *DialService) Update(ctx context.Context, u UpdateDial) error {
	_, err := svc.Get(ctx, u.ID)
	if err != nil {
		return err
	}
	return svc.db.Queries.UpdateDial(ctx, model.UpdateDialParams{
		ID:   u.ID,
		Name: u.Name,
	})
}

type SetDialValue struct {
	ID    int64
	Value int64
}

func (svc *DialService) SetValue(ctx context.Context, v SetDialValue) error {
	_, err := svc.Get(ctx, v.ID)
	if err != nil {
		return err
	}
	return svc.db.Queries.SetDialValue(ctx, model.SetDialValueParams{
		Value: v.Value,
		ID:    v.ID,
	})
}
