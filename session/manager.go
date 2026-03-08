package session

import (
	"context"
	"net/http"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"

	redisdb "subscription_service/db/redis"
)

func NewManager(ctx context.Context) (*scs.SessionManager, *redis.Pool, error) {
	pool, err := redisdb.ConnectPool(ctx)
	if err != nil {
		return nil, nil, err
	}

	manager := scs.New()
	manager.Store = redisstore.New(pool)
	manager.Lifetime = 24 * time.Hour
	manager.Cookie.Name = "subscription_service_session"
	manager.Cookie.HttpOnly = true
	manager.Cookie.Persist = true
	manager.Cookie.SameSite = http.SameSiteLaxMode

	return manager, pool, nil
}
