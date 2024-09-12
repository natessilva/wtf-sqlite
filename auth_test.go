package sqlite_test

import (
	"context"
	"sqlite"
	"testing"
)

func TestAuthServiceSignup(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.CreateAndMigrateDb(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
		return
	}
	svc := sqlite.NewAuthService(db)

	// sign up with new creds
	output, err := svc.Signup(ctx, sqlite.AuthInput{
		UserName: "test",
		Password: "test",
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	if !output.OK {
		t.Fatalf("expected successful signup")
		return
	}
	userID, err := svc.GetUserFromSession(ctx, output.Token)
	if err != nil {
		t.Fatal(err)
		return
	}
	if userID == 0 {
		t.Fatalf("expected non-zero user, got 0")
		return
	}

	// sign up with the existing creds
	output, err = svc.Signup(ctx, sqlite.AuthInput{
		UserName: "test",
		Password: "test",
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	if output.OK {
		t.Fatalf("expected unsuccessful signup")
		return
	}
}

func TestAuthServiceLogin(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.CreateAndMigrateDb(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
		return
	}
	svc := sqlite.NewAuthService(db)

	svc.Signup(ctx, sqlite.AuthInput{
		UserName: "test",
		Password: "test",
	})

	// login with the correct creds
	output, err := svc.Login(ctx, sqlite.AuthInput{
		UserName: "test",
		Password: "test",
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	if !output.OK {
		t.Fatal("expected successful login")
	}

	// login with the incorrect password
	output, err = svc.Login(ctx, sqlite.AuthInput{
		UserName: "test",
		Password: "test wrong",
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	if output.OK {
		t.Fatal("expected failed login")
	}

	// login with an invalid username
	output, err = svc.Login(ctx, sqlite.AuthInput{
		UserName: "test wrong",
		Password: "test",
	})
	if err != nil {
		t.Fatal(err)
		return
	}
	if output.OK {
		t.Fatal("expected failed login")
	}
}
