package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	createuser "edu-portal/app/k8s/create_user"
	"edu-portal/app/store"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var (
	userId = flag.Int("user-id", -1, "user id")
)

func main() {
	flag.Parse()
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

	user, err := store.
		New(db).
		GetUserById(context.Background(), *userId)
	if err != nil {
		log.Fatalf("get user: %v", err)
	}

	kubecfg, err := createuser.
		New().
		Create(context.Background(), user)
	if err != nil {
		log.Fatalf("create k8s user: %v", err)
	}
	fmt.Println(kubecfg)
}
