package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/cache"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/config"
	"github.com/soshyant-joshaghani/go-svelte/internal/core/jobs"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	rdb := redis.NewClient(cache.Options(cfg.RedisHost, cfg.RedisPort, cfg.RedisDB, cfg.RedisPassword))
	defer rdb.Close()
	jobs.RunWorker(ctx, rdb)
}
