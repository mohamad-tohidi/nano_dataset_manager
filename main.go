package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	port := getEnv("PORT", "8080")
	datasetsDir := getEnv("DATASETS_DIR", "./data/datasets")
	dbPath := getEnv("DATABASE_PATH", "./data/metadata.db")

	os.MkdirAll(datasetsDir, 0755)
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	store, err := NewStore(dbPath)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer store.Close()

	srv := NewServer(store, datasetsDir)

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
