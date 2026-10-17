// Command showbuzz runs one of the showbuzz backend services:
//
//	showbuzz seed       migrate the schema, seed the catalog, generate shows
//	showbuzz catalog    discovery, details, search, images     (:8081)
//	showbuzz showtimes  showtimes, cinemas, seat-layout primary (:8082)
//	showbuzz trans      seat locks, transactions, bookings     (:8083)
//	showbuzz user       OTP sign-in and profiles               (:8084)
//	showbuzz notify     Kafka consumer that sends e-mails
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NikhilBhatt-15/showbuzz/services/internal/catalog"
	"github.com/NikhilBhatt-15/showbuzz/services/internal/data"
	"github.com/NikhilBhatt-15/showbuzz/services/internal/notify"
	"github.com/NikhilBhatt-15/showbuzz/services/internal/platform"
	"github.com/NikhilBhatt-15/showbuzz/services/internal/showtimes"
	"github.com/NikhilBhatt-15/showbuzz/services/internal/trans"
	"github.com/NikhilBhatt-15/showbuzz/services/internal/user"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: showbuzz <seed|catalog|showtimes|trans|user|notify>")
		os.Exit(2)
	}
	svc := os.Args[1]
	platform.Logger(svc)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	addr := ":" + platform.Env("PORT", "8080")
	switch svc {
	case "seed":
		db := platform.MustPostgres(ctx)
		defer db.Close()
		if err := data.Bootstrap(ctx, db, platform.Today()); err != nil {
			slog.Error("seed failed", "err", err)
			os.Exit(1)
		}

	case "catalog":
		db, rdb, pub := platform.MustPostgres(ctx), platform.MustRedis(ctx), mustProducer()
		defer pub.Close()
		r := platform.NewRouter()
		catalog.New(db, rdb, pub).Routes(r)
		platform.Serve(addr, r)

	case "showtimes":
		db, rdb := platform.MustPostgres(ctx), platform.MustRedis(ctx)
		go keepShowsFresh(ctx, db)
		r := platform.NewRouter()
		showtimes.New(db, rdb).Routes(r)
		platform.Serve(addr, r)

	case "trans":
		db, rdb, pub := platform.MustPostgres(ctx), platform.MustRedis(ctx), mustProducer()
		defer pub.Close()
		r := platform.NewRouter()
		trans.New(db, rdb, pub).Routes(r)
		platform.Serve(addr, r)

	case "user":
		db, rdb, pub := platform.MustPostgres(ctx), platform.MustRedis(ctx), mustProducer()
		defer pub.Close()
		r := platform.NewRouter()
		user.New(db, rdb, pub).Routes(r)
		platform.Serve(addr, r)

	case "notify":
		for {
			err := notify.New().Run(ctx)
			if ctx.Err() != nil {
				return
			}
			slog.Error("consumer stopped, restarting", "err", err)
			time.Sleep(3 * time.Second)
		}

	default:
		fmt.Fprintln(os.Stderr, "unknown service:", svc)
		os.Exit(2)
	}
}

func mustProducer() *platform.Producer {
	p, err := platform.NewProducer()
	if err != nil {
		slog.Error("kafka producer", "err", err)
		os.Exit(1)
	}
	return p
}

// keepShowsFresh extends the rolling seven-day show window every hour, so a
// long-running stack always has bookable shows.
func keepShowsFresh(ctx context.Context, db *pgxpool.Pool) {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if n, err := data.EnsureShows(ctx, db, platform.Today(), 7); err != nil {
				slog.Error("extend show window", "err", err)
			} else if n > 0 {
				slog.Info("extended show window", "created", n)
			}
		}
	}
}
