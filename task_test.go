package sqlite_test

import (
	"context"
	"fmt"
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

	taskCreated, err := svc.Create(ctx, "test", "description")
	if err != nil {
		t.Fatal(err)
		return
	}
	task, err := svc.Get(ctx, taskCreated.ID)
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
	updatedTask, err := svc.Get(ctx, taskCreated.ID)
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

	err = svc.Delete(ctx, taskCreated.ID)
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

func TestTaskServiceReorderign(t *testing.T) {
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

	ctx = sqlite.ContextWithUser(ctx, model.User{ID: id})

	// create 3 tasks
	for i := 1; i <= 3; i++ {
		_, err := svc.Create(ctx, fmt.Sprintf("task %d", i), fmt.Sprintf("description %d", i))
		if err != nil {
			t.Fatal(err)
			return
		}
	}

	tasks, err := svc.List(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}
	for i, task := range tasks {
		if task.Title != fmt.Sprintf("task %d", i+1) {
			t.Fatalf("expected task %d to have title %q, got %q", i+1, fmt.Sprintf("task %d", i+1), task.Title)
		}
	}

	err = svc.InsertBefore(ctx, tasks[2].ID, tasks[0].ID)
	if err != nil {
		t.Fatal(err)
		return
	}
	newTasks, err := svc.List(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(newTasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(newTasks))
	}
	if newTasks[0].ID != tasks[2].ID {
		t.Fatalf("expected first task to be ID %d, got %d", tasks[2].ID, newTasks[0].ID)
	}
	if newTasks[1].ID != tasks[0].ID {
		t.Fatalf("expected second task to be ID %d, got %d", tasks[1].ID, newTasks[1].ID)
	}
	if newTasks[2].ID != tasks[1].ID {
		t.Fatalf("expected third task to be ID %d, got %d", tasks[0].ID, newTasks[2].ID)
	}

	err = svc.InsertAtEnd(ctx, tasks[0].ID)
	if err != nil {
		t.Fatal(err)
		return
	}
	newTasks, err = svc.List(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(newTasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(newTasks))
	}
	if newTasks[0].ID != tasks[2].ID {
		t.Fatalf("expected first task to be ID %d, got %d", tasks[2].ID, newTasks[0].ID)
	}
	if newTasks[1].ID != tasks[1].ID {
		t.Fatalf("expected second task to be ID %d, got %d", tasks[1].ID, newTasks[1].ID)
	}
	if newTasks[2].ID != tasks[0].ID {
		t.Fatalf("expected third task to be ID %d, got %d", tasks[0].ID, newTasks[2].ID)
	}
}
