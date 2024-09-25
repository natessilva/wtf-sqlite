package sqlite_test

import (
	"context"
	"database/sql"
	"sqlite"
	"sqlite/model"
	"testing"
)

func TestDialService(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.CreateAndMigrateDb(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
		return
	}
	svc := sqlite.NewDialService(db)

	id, err := db.Queries.CreateUser(ctx, model.CreateUserParams{
		UserName: "foo",
		Password: []byte("foo"),
	})
	if err != nil {
		t.Fatal(err)
		return
	}

	// set the logged in user
	ctx = sqlite.ContextWithUser(ctx, model.TeamUser{UserID: id})

	dials, err := svc.List(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(dials) != 0 {
		t.Fatalf("expected zero dials, got %d", len(dials))
	}

	dialId, err := svc.Create(ctx, "test")
	if err != nil {
		t.Fatal(err)
		return
	}
	dial, err := svc.Get(ctx, dialId)
	if err != nil {
		t.Fatal(err)
		return
	}
	if dial.Name != "test" {
		t.Fatalf("expected name test, got %s", dial.Name)
	}
	dials, err = svc.List(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(dials) != 1 {
		t.Fatalf("expected one dial, got %d", len(dials))
	}

	// create another user
	id2, err := db.Queries.CreateUser(ctx, model.CreateUserParams{
		UserName: "bar",
		Password: []byte("foo"),
	})
	if err != nil {
		t.Fatal(err)
		return
	}

	// log in the second user
	ctx = sqlite.ContextWithUser(ctx, model.TeamUser{UserID: id2})
	dials, err = svc.List(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(dials) != 0 {
		t.Fatalf("expected zero dials, got %d", len(dials))
	}

	dialId2, err := svc.Create(ctx, "test")
	if err != nil {
		t.Fatal(err)
		return
	}
	dial, err = svc.Get(ctx, dialId2)
	if err != nil {
		t.Fatal(err)
		return
	}
	if dial.Name != "test" {
		t.Fatalf("expected name test, got %s", dial.Name)
	}
	dials, err = svc.List(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(dials) != 1 {
		t.Fatalf("expected one dial, got %d", len(dials))
	}

	// second user cannot see first user's ids
	_, err = svc.Get(ctx, dialId)
	if err != sql.ErrNoRows {
		t.Fatal("expected rows we don't have access to to be invisible")
	}

	// first user cannot see second user's ids
	ctx = sqlite.ContextWithUser(ctx, model.TeamUser{UserID: id})
	_, err = svc.Get(ctx, dialId2)
	if err != sql.ErrNoRows {
		t.Fatal("expected rows we don't have access to to be invisible")
	}
}
