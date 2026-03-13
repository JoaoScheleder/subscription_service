package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"subscription_service/db/postgres"
	data "subscription_service/models"
	"subscription_service/session"
	"sync"
	"syscall"
)

const PORT = "8080"

func main() {
	// connect to database

	ctx := context.Background()

	pool, err := postgres.ConnectPool(ctx)
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer pool.Close()

	// create loggers
	infoLog := log.New(log.Writer(), "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(log.Writer(), "ERROR\t", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)

	// create sessions
	sessionManager, redisPool, err := session.NewManager(ctx)
	if err != nil {
		log.Fatalf("create session manager: %v", err)
	}
	defer redisPool.Close()
	log.Printf("session manager initialized (lifetime=%s)", sessionManager.Lifetime)

	// create channels

	// create waitgroup
	wg := &sync.WaitGroup{}

	// set up the application config

	app := Config{
		Session:   sessionManager,
		DB:        pool,
		InfoLog:   infoLog,
		ErrorLog:  errorLog,
		WaitGroup: wg,
		Models:    data.New(pool),
	}

	// set up mail

	// listen for connections
	app.serve()
}

func (app *Config) serve() {
	srv := &http.Server{
		Addr:    ":" + PORT,
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}

func (app *Config) ListenForShutdown() {
	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	app.shutdown()
	os.Exit(0)
}

func (app *Config) shutdown() {
	// perform any necessary cleanup here (e.g. close database connections, stop background workers, etc.)
	app.InfoLog.Println("shutting down server...")

	app.WaitGroup.Wait()

	app.InfoLog.Println("server stopped")
}
