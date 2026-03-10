package main

import (
	"context"
	"log"
	"net/http"
	"subscription_service/db/postgres"
	"subscription_service/session"
	"sync"
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
