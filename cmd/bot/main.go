package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	edu_bot "edu-portal/app/bot"
	"edu-portal/app/store"

	"github.com/go-telegram/bot"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	godotenv.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	sqliteDB, err := sql.Open("sqlite3", "./db.sqlite")
	if err != nil {
		log.Fatalf("init db: %v", err)
	}

	db := sqlx.NewDb(sqliteDB, "sqlite3")
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	handler := edu_bot.New(store.New(db))
	tgToken, exists := os.LookupEnv("TELEGRAM_TOKEN")
	if !exists {
		log.Fatalf("env var TELEGRAM_TOKEN is required")
	}

	tgBot, err := bot.New(tgToken, bot.WithDefaultHandler(handler.Handle))
	if err != nil {
		log.Fatalf("create tg bot: %v", err)
	}
	tgBot.Start(ctx)
}
