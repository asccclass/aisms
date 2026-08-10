/*
   ISMS資訊資產管理系統
   後端使用 sherryserver 架構設計（標準 net/http）
*/
package main

import (
	"fmt"
	"isms-privilege/internal/db"
	"isms-privilege/internal/handlers"
	"isms-privilege/internal/mailer"
	"log"
	"net/http"
	"os"

	SherryServer "github.com/asccclass/sherryserver"
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

	// 確保資料目錄存在
	os.MkdirAll("data", 0755)
	os.MkdirAll("logs", 0755)

	database, err := db.New(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()
	log.Printf("[DB] Initialized: %s", dbPath)

	mail := mailer.New()
	h := handlers.New(database, mail)
	handlers.InitOAuth2()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	docRoot := os.Getenv("DocumentRoot")
	if docRoot == "" {
		docRoot = os.Getenv("DOCUMENT_ROOT")
	}
	if docRoot == "" {
		docRoot = "www/html"
	}

	templateRoot := os.Getenv("TemplateRoot")
	if templateRoot == "" {
		templateRoot = "www/template"
	}
	mux.Handle("/", http.FileServer(http.Dir(docRoot)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv, err := SherryServer.NewServer(":"+port, docRoot, templateRoot)
	if err != nil {
		log.Fatalf("failed to create sherryserver: %v", err)
	}
	srv.Server.Handler = corsMiddleware(mux)

	fmt.Printf("\n🔐 ISMS資訊資產管理系統\n")
	fmt.Printf("   伺服器位址：http://localhost:%s\n", port)
	fmt.Printf("   靜態檔案：%s\n\n", docRoot)

	srv.Start()
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := os.Getenv("CORS_ORIGIN")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
