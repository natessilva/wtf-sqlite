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
	dialService := sqlite.NewDialService(db)
	var server *http.Server

	env := os.Getenv("ENV")

	if env == "prod" {
		certManager := autocert.Manager{
			Cache:      autocert.DirCache("certs"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("silva.world"),
		}

		server = &http.Server{
			Addr: ":443",
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
			Handler: sqlite.NewHandler(authService, userService, dialService, true),
		}
		go func() { http.ListenAndServe(":80", certManager.HTTPHandler(nil)) }()
		go func() { log.Fatal(server.ListenAndServeTLS("", "")) }()
		log.Println("server running on ports 80 and 443")
	} else {

		server = &http.Server{
			Addr:    ":8000",
			Handler: sqlite.NewHandler(authService, userService, dialService, false),
		}

		go func() { log.Fatal(server.ListenAndServe()) }()
		log.Println("server running on port 8000")
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	go func() { log.Fatal(http.ListenAndServe(":6060", mux)) }()

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
