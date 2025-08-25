package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sqlite"
	"time"

	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/acme/autocert"
)

func run() error {
	ctx := context.Background()
	db, err := sqlite.CreateAndMigrateDb(ctx, "db/app.db")
	if err != nil {
		return err
	}
	defer db.Close()

	authService := sqlite.NewAuthService(db)
	userService := sqlite.NewUserService(db)
	taskService := sqlite.NewTaskService(db)

	sessionCleanupService := sqlite.NewSessionCleanupService(db)
	defer sessionCleanupService.Close()

	var server *http.Server

	env := os.Getenv("ENV")

	handler := sqlite.NewHandler(authService, userService, taskService, env == "prod")

	if env == "prod" {
		server = startProdServer(handler)
	} else {
		server = startLocalServer(handler)
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

func startLocalServer(handler http.Handler) *http.Server {
	server := &http.Server{
		Addr:    ":8000",
		Handler: handler,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	log.Println("server running on port 8000")
	return server
}

func startProdServer(handler http.Handler) *http.Server {
	certManager := autocert.Manager{
		Cache:      autocert.DirCache("certs"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("silva.world"),
	}

	server := &http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
		Handler: handler,
	}
	go func() {
		if err := http.ListenAndServe(":80", certManager.HTTPHandler(nil)); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	log.Println("server running on ports 80 and 443")
	return server
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
