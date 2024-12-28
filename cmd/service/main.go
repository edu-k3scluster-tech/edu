package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"edu-portal/app/server"
	"edu-portal/app/store"
	"edu-portal/app/store/migrator"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	godotenv.Load()

	dbPath, exists := os.LookupEnv("DB_PATH")
	if !exists {
		log.Fatalf("DB_PATH env is required")
	}

	sqliteDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("init db: %v", err)
	}

	db := sqlx.NewDb(sqliteDB, "sqlite3")
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	migrator := migrator.New(db)
	if err := migrator.Run(context.Background()); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	tpls, err := parseTemplates()
	if err != nil {
		log.Fatalf("collect templates: %v", err)
	}

	srv := server.New(false, tpls, store.New(db))
	if err := srv.Run(context.Background()); err != nil {
		log.Fatalf("run srv: %v", err)
	}
}
