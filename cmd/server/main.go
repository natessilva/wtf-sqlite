package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sqlite"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func run() error {
	ctx := context.Background()
	db, err := sqlite.CreateAndMigrateDb(ctx, "db/app.db")
	if err != nil {
		return err
	}
	defer db.Close()

	// todo application server
	authService := sqlite.NewAuthService(db)
	userService := sqlite.NewUserService(db)
	dialService := sqlite.NewDialService(db)

	server := &http.Server{
		Addr:    ":8000",
		Handler: sqlite.NewHandler(authService, userService, dialService),
	}

	go log.Fatal(server.ListenAndServe())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
