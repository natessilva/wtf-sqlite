package sqlite

import (
	"context"
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
	return svc.db.Queries.CreateTask(ctx, model.CreateTaskParams{
		UserID:      UserFromContext(ctx).ID,
		Title:       title,
		Description: description,
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
