package sqlite_test

import (
	"context"
	"fmt"
	"sqlite"
	"sqlite/model"
	"testing"
)

func TestUserService(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.CreateAndMigrateDb(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
		return
	}
	svc := sqlite.NewUserService(db)

	// do not set any authenticated user at all
	_, err = svc.Get(ctx)
	if err == nil {
		t.Fatalf("expected error")
	}
	// set the authenticated userID to one that doesn't exist
	ctx = sqlite.ContextWithUser(ctx, 1)
	_, err = svc.Get(ctx)
	if err == nil {
		t.Fatalf("expected error")
	}

	// create a real user
	id, err := db.Queries.CreateUser(ctx, model.CreateUserParams{
		UserName: "test",
		Password: []byte("hash"),
	})
	if err != nil {
		t.Fatal(err)
		return
	}

	// set the authenticated user to the existing user
	ctx = sqlite.ContextWithUser(ctx, id)
	user, err := svc.Get(ctx)
	if err != nil {
		t.Fatal(fmt.Errorf("expected no error, got %w", err))
	}
	if user.UserName != "test" {
		t.Fatalf("expected username test, got %s", user.UserName)
	}
}
