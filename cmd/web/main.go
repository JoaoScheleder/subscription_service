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

	mail "github.com/xhit/go-simple-mail/v2"
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
		Session:       sessionManager,
		DB:            pool,
		InfoLog:       infoLog,
		ErrorLog:      errorLog,
		WaitGroup:     wg,
		Models:        data.New(pool),
		ErrorChan:     make(chan error),
		ErrorChanDone: make(chan bool),
	}

	// set up mail
	app.Mailer = app.CreateMail()

	go app.ListenForMail()

	go app.ListenForErrors()
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

func (app *Config) ListenForErrors() {
	for {
		select {
		case err := <-app.ErrorChan:
			app.ErrorLog.Printf("error: %v", err)
		case <-app.ErrorChanDone:
			return
		}
	}
}

func (app *Config) shutdown() {
	// perform any necessary cleanup here (e.g. close database connections, stop background workers, etc.)
	app.InfoLog.Println("shutting down server...")

	app.WaitGroup.Wait()

	app.InfoLog.Println("server stopped")

	app.Mailer.DoneChan <- true
	app.ErrorChanDone <- true

	close(app.Mailer.DoneChan)
	close(app.Mailer.MailerChan)
	close(app.ErrorChanDone)
	close(app.ErrorChan)
}

func (app *Config) CreateMail() *Mail {

	mailerChan := make(chan Message)
	doneChan := make(chan bool)

	m := &Mail{
		Domain:      "localhost",
		Host:        "localhost",
		Port:        1025,
		Username:    "",
		Password:    "",
		Encryption:  mail.EncryptionNone,
		FromAddress: "no-reply@example.com",
		FromName:    "Example App",
		Wait:        app.WaitGroup,
		MailerChan:  mailerChan,
		DoneChan:    doneChan,
	}

	return m
}
