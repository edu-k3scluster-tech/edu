package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"edu-portal/app/cluster"
	"edu-portal/app/server"
	"edu-portal/app/store"
	"edu-portal/app/store/migrator"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	godotenv.Load()

	dbPath, exists := os.LookupEnv("DB_PATH")
	if !exists {
		log.Fatalf("DB_PATH env is required")
	}

	var cfg *rest.Config
	var err error

	if path, exists := os.LookupEnv("KUBECONFIG_PATH"); exists {
		log.Printf("[INFO] Get k8s config from %s", path)
		if cfg, err = clientcmd.BuildConfigFromFlags("", path); err != nil {
			log.Fatalf("get kube config from %s", path)
		}
	} else {
		log.Printf("[INFO] Get in-cluster k8s config")
		if cfg, err = rest.InClusterConfig(); err != nil {
			log.Fatalf("get in-cluster kube config")
		}
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("build kube client")
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

	srv := server.New(false, tpls, store.New(db), cluster.New(clientset, cfg.CAData))
	if err := srv.Run(context.Background()); err != nil {
		log.Fatalf("run srv: %v", err)
	}
}
