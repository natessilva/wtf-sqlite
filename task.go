package sqlite

import (
	"context"
	"database/sql"
	"sqlite/model"
)

type TaskService struct {
	db *DB
}

func NewTaskService(db *DB) *TaskService {
	return &TaskService{
		db: db,
	}
}

func (svc *TaskService) Create(ctx context.Context, title, description string) (model.Task, error) {
	max, err := svc.db.Queries.GetMaxTaskOrdinal(ctx, UserFromContext(ctx).ID)
	if err != nil {
		if err != sql.ErrNoRows {
			return model.Task{}, err
		}
		max = 0
	}
	return svc.db.Queries.CreateTask(ctx, model.CreateTaskParams{
		UserID:      UserFromContext(ctx).ID,
		Title:       title,
		Description: description,
		Ordinal:     max + 100,
	})
}

func (svc *TaskService) List(ctx context.Context) ([]model.Task, error) {
	return svc.db.Queries.ListTasks(ctx, UserFromContext(ctx).ID)
}

func (svc *TaskService) Get(ctx context.Context, id int64) (model.Task, error) {
	return svc.db.Queries.GetTask(ctx, model.GetTaskParams{
		UserID: UserFromContext(ctx).ID,
		ID:     id,
	})
}

func (svc *TaskService) Update(ctx context.Context, t model.Task) error {
	_, err := svc.Get(ctx, t.ID)
	if err != nil {
		return err
	}
	return svc.db.Queries.UpdateTask(ctx, model.UpdateTaskParams{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		IsCompleted: t.IsCompleted,
	})
}

func (svc *TaskService) Delete(ctx context.Context, id int64) error {
	_, err := svc.Get(ctx, id)
	if err != nil {
		return err
	}
	return svc.db.Queries.DeleteTask(ctx, id)
}

func (svc *TaskService) InsertBefore(ctx context.Context, idToInsert, target int64) error {
	_, err := svc.Get(ctx, idToInsert)
	if err != nil {
		return err
	}
	t, err := svc.Get(ctx, target)
	if err != nil {
		return err
	}
	previousOrdinal, err := svc.db.Queries.GetPreviousTaskOrdinal(ctx, model.GetPreviousTaskOrdinalParams{
		UserID:  UserFromContext(ctx).ID,
		Ordinal: t.Ordinal,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		previousOrdinal = t.Ordinal - 100
	}
	midpoint := (previousOrdinal + t.Ordinal) / 2
	return svc.db.Queries.SetTaskOrdinal(ctx, model.SetTaskOrdinalParams{
		Ordinal: midpoint,
		ID:      idToInsert,
	})
}

func (svc *TaskService) InsertAtEnd(ctx context.Context, idToInsert int64) error {
	_, err := svc.Get(ctx, idToInsert)
	if err != nil {
		return err
	}
	maxOrdinal, err := svc.db.Queries.GetMaxTaskOrdinal(ctx, UserFromContext(ctx).ID)
	if err != nil {
		return err
	}
	return svc.db.Queries.SetTaskOrdinal(ctx, model.SetTaskOrdinalParams{
		Ordinal: maxOrdinal + 100,
		ID:      idToInsert,
	})
}
