package main

import (
	"isms-privilege/internal/db"
	"isms-privilege/internal/mcp"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("envfile"); err != nil {
		log.Println("Warning: envfile not found, using environment variables")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/isms.db"
	}

	database, err := db.New(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if err := mcp.NewServer(database).ServeStdio(os.Stdin, os.Stdout); err != nil {
		log.Fatalf("mcp stdio server failed: %v", err)
	}
}
