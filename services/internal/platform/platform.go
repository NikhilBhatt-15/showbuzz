// Package platform holds the plumbing shared by every showbuzz service:
// configuration, Postgres/Redis/Kafka clients, HTTP helpers and auth.
package platform

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// IST is India Standard Time. Show dates and times are always IST, the same
// convention BookMyShow uses for its yyyymmdd "dateCode" values.
var IST = time.FixedZone("IST", 5*3600+1800)

func Env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func EnvInt(key string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}
	return def
}

func EnvBool(key string) bool {
	v, _ := strconv.ParseBool(os.Getenv(key))
	return v
}

// Logger returns a JSON logger tagged with the service name.
func Logger(service string) *slog.Logger {
	l := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", service)
	slog.SetDefault(l)
	return l
}

// MustPostgres connects to DATABASE_URL, retrying while the database boots.
func MustPostgres(ctx context.Context) *pgxpool.Pool {
	url := Env("DATABASE_URL", "postgres://showbuzz:showbuzz@localhost:5432/showbuzz")
	for attempt := 1; ; attempt++ {
		pool, err := pgxpool.New(ctx, url)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				return pool
			}
			pool.Close()
		}
		if attempt >= 30 {
			slog.Error("postgres unavailable", "err", err)
			os.Exit(1)
		}
		slog.Warn("waiting for postgres", "attempt", attempt, "err", err)
		time.Sleep(2 * time.Second)
	}
}

// MustRedis connects to REDIS_ADDR, retrying while Redis boots.
func MustRedis(ctx context.Context) *redis.Client {
	rdb := redis.NewClient(&redis.Options{Addr: Env("REDIS_ADDR", "localhost:6379")})
	for attempt := 1; ; attempt++ {
		err := rdb.Ping(ctx).Err()
		if err == nil {
			return rdb
		}
		if attempt >= 30 {
			slog.Error("redis unavailable", "err", err)
			os.Exit(1)
		}
		slog.Warn("waiting for redis", "attempt", attempt, "err", err)
		time.Sleep(2 * time.Second)
	}
}

// DateCode formats t as the yyyymmdd code used in URLs and APIs.
func DateCode(t time.Time) string { return t.In(IST).Format("20060102") }

// ParseDateCode parses a yyyymmdd code as midnight IST.
func ParseDateCode(code string) (time.Time, error) {
	return time.ParseInLocation("20060102", code, IST)
}

// Today returns midnight IST of the current day.
func Today() time.Time {
	now := time.Now().In(IST)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, IST)
}
