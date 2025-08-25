package sqlite_test

import (
	"context"
	"sqlite"
	"sqlite/model"
	"testing"
)

func TestTaskService(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.CreateAndMigrateDb(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
		return
	}
	svc := sqlite.NewTaskService(db)

	id, err := db.Queries.CreateUser(ctx, model.CreateUserParams{
		UserName: "foo",
		Password: []byte("foo"),
	})
	if err != nil {
		t.Fatal(err)
		return
	}

	// set the logged in user
	ctx = sqlite.ContextWithUser(ctx, model.User{ID: id})

	tasks, err := svc.List(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(tasks) != 0 {
		t.Fatalf("expected zero tasks, got %d", len(tasks))
	}

	taskId, err := svc.Create(ctx, "test", "description")
	if err != nil {
		t.Fatal(err)
		return
	}
	task, err := svc.Get(ctx, taskId)
	if err != nil {
		t.Fatal(err)
		return
	}
	if task.Title != "test" {
		t.Fatalf("expected name test, got %s", task.Title)
	}
	tasks, err = svc.List(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(tasks) != 1 {
		t.Fatalf("expected one task, got %d", len(tasks))
	}

	task.Title = "updated"
	task.Description = "updated description"
	task.IsCompleted = true
	err = svc.Update(ctx, task)
	if err != nil {
		t.Fatal(err)
		return
	}
	updatedTask, err := svc.Get(ctx, taskId)
	if err != nil {
		t.Fatal(err)
		return
	}
	if updatedTask.Title != "updated" {
		t.Fatalf("expected name updated, got %s", updatedTask.Title)
	}
	if updatedTask.Description != "updated description" {
		t.Fatalf("expected description updated description, got %s", updatedTask.Description)
	}
	if !updatedTask.IsCompleted {
		t.Fatalf("expected completed true, got false")
	}

	err = svc.Delete(ctx, taskId)
	if err != nil {
		t.Fatal(err)
		return
	}
	tasks, err = svc.List(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(tasks) != 0 {
		t.Fatalf("expected zero tasks, got %d", len(tasks))
	}
}
