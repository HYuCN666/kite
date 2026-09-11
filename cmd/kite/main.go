package main

import (
	"log"

	"github.com/HYuCN666/kite/internal/config"
	"github.com/HYuCN666/kite/internal/server"
	"github.com/HYuCN666/kite/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := store.Open(cfg.DataDir)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer db.Close()

	srv := server.New(cfg, db)
	if err := srv.Run(); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
