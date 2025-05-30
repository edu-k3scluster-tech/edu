package main

import (
	"context"
	"database/sql"
	"encoding/base64"
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

	if err := rest.LoadTLSFiles(cfg); err != nil {
		log.Fatalf("load tls files")
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

	privateKeyB64, exists := os.LookupEnv("JWT_PRIVATE_KEY")
	if !exists {
		log.Fatalf("JWT_PRIVATE_KEY env is required")
	}

	publicKeyB64, exists := os.LookupEnv("JWT_PUBLIC_KEY")
	if !exists {
		log.Fatalf("JWT_PUBLIC_KEY env is required")
	}

	privateKey, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		log.Fatalf("decode JWT_PRIVATE_KEY: %v", err)
	}

	publicKey, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		log.Fatalf("decode JWT_PUBLIC_KEY: %v", err)
	}
	srv := server.New(
		false,
		tpls,
		store.New(db),
		cluster.New(clientset, cfg.CAData),
		privateKey,
		publicKey,
	)
	if err := srv.Run(context.Background()); err != nil {
		log.Fatalf("run srv: %v", err)
	}
}
