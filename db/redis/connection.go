package redisdb

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	"subscription_service/config/env"
)

type Config struct {
	Addr           string
	Password       string
	DB             int
	MaxIdle        int
	MaxActive      int
	IdleTimeout    time.Duration
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	Wait           bool
}

func DefaultConfig() Config {
	return Config{
		Addr:           "localhost:6379",
		Password:       "",
		DB:             0,
		MaxIdle:        10,
		MaxActive:      50,
		IdleTimeout:    5 * time.Minute,
		ConnectTimeout: 2 * time.Second,
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		Wait:           true,
	}
}

func LoadConfigFromEnv() (Config, error) {
	cfg := DefaultConfig()

	cfg.Addr = env.String("REDIS_ADDR", cfg.Addr)
	cfg.Password = env.String("REDIS_PASSWORD", cfg.Password)

	db, err := env.Int("REDIS_DB", cfg.DB)
	if err != nil {
		return Config{}, err
	}
	cfg.DB = db

	maxIdle, err := env.Int("REDIS_MAX_IDLE", cfg.MaxIdle)
	if err != nil {
		return Config{}, err
	}
	cfg.MaxIdle = maxIdle

	maxActive, err := env.Int("REDIS_MAX_ACTIVE", cfg.MaxActive)
	if err != nil {
		return Config{}, err
	}
	cfg.MaxActive = maxActive

	idleTimeout, err := env.Duration("REDIS_IDLE_TIMEOUT", cfg.IdleTimeout)
	if err != nil {
		return Config{}, err
	}
	cfg.IdleTimeout = idleTimeout

	connectTimeout, err := env.Duration("REDIS_CONNECT_TIMEOUT", cfg.ConnectTimeout)
	if err != nil {
		return Config{}, err
	}
	cfg.ConnectTimeout = connectTimeout

	readTimeout, err := env.Duration("REDIS_READ_TIMEOUT", cfg.ReadTimeout)
	if err != nil {
		return Config{}, err
	}
	cfg.ReadTimeout = readTimeout

	writeTimeout, err := env.Duration("REDIS_WRITE_TIMEOUT", cfg.WriteTimeout)
	if err != nil {
		return Config{}, err
	}
	cfg.WriteTimeout = writeTimeout

	wait, err := env.Bool("REDIS_WAIT", cfg.Wait)
	if err != nil {
		return Config{}, err
	}
	cfg.Wait = wait

	return cfg, nil
}

func ConnectPool(ctx context.Context) (*redis.Pool, error) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	pool := NewPool(cfg)

	_ = ctx

	conn := pool.Get()
	if err := conn.Err(); err != nil {
		pool.Close()
		return nil, fmt.Errorf("acquire redis connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.Do("PING"); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return pool, nil
}

func NewPool(cfg Config) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     cfg.MaxIdle,
		MaxActive:   cfg.MaxActive,
		IdleTimeout: cfg.IdleTimeout,
		Wait:        cfg.Wait,
		Dial: func() (redis.Conn, error) {
			opts := []redis.DialOption{
				redis.DialConnectTimeout(cfg.ConnectTimeout),
				redis.DialReadTimeout(cfg.ReadTimeout),
				redis.DialWriteTimeout(cfg.WriteTimeout),
				redis.DialDatabase(cfg.DB),
			}

			if cfg.Password != "" {
				opts = append(opts, redis.DialPassword(cfg.Password))
			}

			return redis.Dial("tcp", cfg.Addr, opts...)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}
