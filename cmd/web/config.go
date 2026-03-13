package main

import (
	"log"
	"sync"

	data "subscription_service/models"

	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Session   *scs.SessionManager
	DB        *pgxpool.Pool
	InfoLog   *log.Logger
	ErrorLog  *log.Logger
	WaitGroup *sync.WaitGroup
	Models    data.Models
}
