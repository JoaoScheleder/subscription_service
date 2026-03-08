package main

import (
	"context"
	"log"
	"subscription_service/db/postgres"
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

	// create sessions

	// create channels

	// create waitgroup

	// set up the application config

	// set up mail

	// listen for connections
}
