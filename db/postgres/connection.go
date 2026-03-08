package postgres

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"subscription_service/config/env"
)

type Config struct {
	Host              string
	Port              int
	User              string
	Password          string
	Database          string
	SSLMode           string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}

func DefaultConfig() Config {
	return Config{
		Host:              "localhost",
		Port:              5432,
		User:              "postgres",
		Password:          "",
		Database:          "postgres",
		SSLMode:           "disable",
		MaxConns:          25,
		MinConns:          2,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
		ConnectTimeout:    5 * time.Second,
	}
}

func LoadConfigFromEnv() (Config, error) {
	cfg := DefaultConfig()

	cfg.Host = env.String("DB_HOST", cfg.Host)
	cfg.User = env.String("DB_USER", cfg.User)
	cfg.Password = env.String("DB_PASSWORD", cfg.Password)
	cfg.Database = env.String("DB_NAME", cfg.Database)
	cfg.SSLMode = env.String("DB_SSLMODE", cfg.SSLMode)

	port, err := env.Int("DB_PORT", cfg.Port)
	if err != nil {
		return Config{}, err
	}
	cfg.Port = port

	maxConns, err := env.Int("DB_MAX_CONNS", int(cfg.MaxConns))
	if err != nil {
		return Config{}, err
	}
	cfg.MaxConns = int32(maxConns)

	minConns, err := env.Int("DB_MIN_CONNS", int(cfg.MinConns))
	if err != nil {
		return Config{}, err
	}
	cfg.MinConns = int32(minConns)

	if cfg.MinConns > cfg.MaxConns {
		return Config{}, fmt.Errorf("invalid pool config: DB_MIN_CONNS (%d) cannot be greater than DB_MAX_CONNS (%d)", cfg.MinConns, cfg.MaxConns)
	}

	maxConnLifetime, err := env.Duration("DB_MAX_CONN_LIFETIME", cfg.MaxConnLifetime)
	if err != nil {
		return Config{}, err
	}
	cfg.MaxConnLifetime = maxConnLifetime

	maxConnIdleTime, err := env.Duration("DB_MAX_CONN_IDLE_TIME", cfg.MaxConnIdleTime)
	if err != nil {
		return Config{}, err
	}
	cfg.MaxConnIdleTime = maxConnIdleTime

	healthCheckPeriod, err := env.Duration("DB_HEALTHCHECK_PERIOD", cfg.HealthCheckPeriod)
	if err != nil {
		return Config{}, err
	}
	cfg.HealthCheckPeriod = healthCheckPeriod

	connectTimeout, err := env.Duration("DB_CONNECT_TIMEOUT", cfg.ConnectTimeout)
	if err != nil {
		return Config{}, err
	}
	cfg.ConnectTimeout = connectTimeout

	return cfg, nil
}

func (c Config) DSN() string {
	query := url.Values{}
	query.Set("sslmode", c.SSLMode)
	query.Set("connect_timeout", strconv.Itoa(int(c.ConnectTimeout.Seconds())))

	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:     c.Database,
		RawQuery: query.Encode(),
	}

	return u.String()
}

func ConnectPool(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	return NewPool(ctx, cfg)
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod

	connectCtx := ctx
	cancel := func() {}
	if cfg.ConnectTimeout > 0 {
		connectCtx, cancel = context.WithTimeout(ctx, cfg.ConnectTimeout)
	}
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres via pool: %w", err)
	}

	return pool, nil
}

func ConnectConn(ctx context.Context) (*pgx.Conn, error) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	connCfg, err := pgx.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	connectCtx := ctx
	cancel := func() {}
	if cfg.ConnectTimeout > 0 {
		connectCtx, cancel = context.WithTimeout(ctx, cfg.ConnectTimeout)
	}
	defer cancel()

	conn, err := pgx.ConnectConfig(connectCtx, connCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	if err := conn.Ping(connectCtx); err != nil {
		conn.Close(ctx)
		return nil, fmt.Errorf("ping postgres connection: %w", err)
	}

	return conn, nil
}
