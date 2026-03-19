package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/pha06ntom/tg-translator-bot/internal/config"
	"github.com/pha06ntom/tg-translator-bot/internal/db"
	openaiwrap "github.com/pha06ntom/tg-translator-bot/internal/openai"
	"github.com/pha06ntom/tg-translator-bot/internal/telegram"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config yaml")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := db.ApplyMigrations(ctx, pool, "migrations"); err != nil {
		log.Fatal(err)
	}

	oa := openaiwrap.New(cfg.OpenAIAPIKey, cfg.OpenAIModel)

	b, err := telegram.New(cfg.TelegramToken, pool, cfg.AdminUserIDs, oa)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("bot started")
	if err := b.Run(ctx); err != nil {
		log.Fatal(err)
	}

}
