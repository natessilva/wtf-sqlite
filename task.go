package sqlite

import (
	"context"
	"database/sql"
	"fmt"
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

func (svc *TaskService) Create(ctx context.Context, title, description string, parentID sql.NullInt64) (model.Task, error) {
	var max float64
	var err error
	if parentID.Valid {
		_, err = svc.Get(ctx, parentID.Int64)
		if err != nil {
			return model.Task{}, fmt.Errorf("parent task not found: %w", err)
		}
		max, err = svc.db.Queries.GetMaxTaskOrdinalByParent(ctx, parentID)
	} else {
		max, err = svc.db.Queries.GetMaxTaskOrdinal(ctx, UserFromContext(ctx).ID)
	}
	if err != nil {
		if err != sql.ErrNoRows {
			return model.Task{}, fmt.Errorf("error getting max ordinal: %w", err)
		}
		max = 0
	}
	return svc.db.Queries.CreateTask(ctx, model.CreateTaskParams{
		UserID:      UserFromContext(ctx).ID,
		Title:       title,
		Description: description,
		Ordinal:     max + 100,
		ParentID:    parentID,
	})
}

func (svc *TaskService) List(ctx context.Context, parentID sql.NullInt64) ([]model.Task, error) {
	if parentID.Valid {
		return svc.db.Queries.ListTasksByParent(ctx, parentID)
	}
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
		return fmt.Errorf("task not found: %w", err)
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
		return fmt.Errorf("task not found: %w", err)
	}
	return svc.db.Queries.DeleteTask(ctx, id)
}

func (svc *TaskService) InsertBefore(ctx context.Context, idToInsert int64, target sql.NullInt64, parentID sql.NullInt64) error {
	if !target.Valid {
		return svc.insertAtEnd(ctx, idToInsert, parentID)
	}
	_, err := svc.Get(ctx, idToInsert)
	if err != nil {
		return fmt.Errorf("task to insert not found: %w", err)
	}
	t, err := svc.Get(ctx, target.Int64)
	if err != nil {
		return fmt.Errorf("target task not found: %w", err)
	}
	var previousOrdinal float64
	if parentID.Valid {
		_, err := svc.Get(ctx, parentID.Int64)
		if err != nil {
			return fmt.Errorf("parent task not found: %w", err)
		}
		previousOrdinal, err = svc.db.Queries.GetPreviousTaskOrdinalByParent(ctx, model.GetPreviousTaskOrdinalByParentParams{
			ParentID: parentID,
			Ordinal:  t.Ordinal,
		})
	} else {
		previousOrdinal, err = svc.db.Queries.GetPreviousTaskOrdinal(ctx, model.GetPreviousTaskOrdinalParams{
			UserID:  UserFromContext(ctx).ID,
			Ordinal: t.Ordinal,
		})

	}
	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("error getting previous ordinal: %w", err)
		}
		previousOrdinal = t.Ordinal - 100
	}
	midpoint := (previousOrdinal + t.Ordinal) / 2
	return svc.db.Queries.SetTaskOrdinal(ctx, model.SetTaskOrdinalParams{
		Ordinal: midpoint,
		ID:      idToInsert,
	})
}

func (svc *TaskService) insertAtEnd(ctx context.Context, idToInsert int64, parentID sql.NullInt64) error {
	_, err := svc.Get(ctx, idToInsert)
	if err != nil {
		return fmt.Errorf("task to insert not found: %w", err)
	}
	var maxOrdinal float64
	if parentID.Valid {
		_, err := svc.Get(ctx, parentID.Int64)
		if err != nil {
			return fmt.Errorf("parent task not found: %w", err)
		}
		maxOrdinal, err = svc.db.Queries.GetMaxTaskOrdinalByParent(ctx, parentID)
	} else {
		maxOrdinal, err = svc.db.Queries.GetMaxTaskOrdinal(ctx, UserFromContext(ctx).ID)
	}
	if err != nil {
		return fmt.Errorf("error getting max ordinal: %w", err)
	}
	return svc.db.Queries.SetTaskOrdinal(ctx, model.SetTaskOrdinalParams{
		Ordinal: maxOrdinal + 100,
		ID:      idToInsert,
	})
}
