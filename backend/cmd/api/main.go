package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/cache"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/config"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/db"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/jobs"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/migrate"
	"github.com/soshyant-joshaghani/go-svelte/internal/httpserver"
	"github.com/soshyant-joshaghani/go-svelte/internal/modules/apps/sample"
	"github.com/soshyant-joshaghani/go-svelte/internal/modules/base/users"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.PostgresURL())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	dir, err := migrate.Dir(cfg.MigrationsDir)
	if err != nil {
		log.Fatal(err)
	}
	if err := migrate.Up(ctx, pool, dir); err != nil {
		log.Fatal(err)
	}

	rdb := redis.NewClient(cache.Options(cfg.RedisHost, cfg.RedisPort, cfg.RedisDB, cfg.RedisPassword))
	defer rdb.Close()
	server := &httpserver.Server{
		Config: cfg,
		Users:  users.NewPgRepository(pool),
		Notes:  sample.NewPgRepository(pool),
		Cache:  cache.NewRedis(rdb),
		Jobs:   jobs.NewRedisQueue(rdb),
	}
	if err := server.UserService().EnsureFirstSuperuser(ctx, cfg.FirstSuperuser, cfg.FirstSuperuserPassword); err != nil {
		log.Fatal(err)
	}

	httpServer := &http.Server{
		Addr:              net.JoinHostPort(cfg.AppHost, strconv.Itoa(cfg.AppPort)),
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Printf("api listening on %s (docs /docs, scalar /sdoc)", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdown)
}
