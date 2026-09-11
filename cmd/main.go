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
	"isms-privilege/internal/web"
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
	renderer, err := web.NewRenderer(templateRoot)
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	fileServer := http.FileServer(http.Dir(docRoot))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			fileServer.ServeHTTP(w, r)
			return
		}
		if err := renderer.Render(w, "index.gohtml", web.PageData{
			Title:    "ISMS資訊資產管理系統",
			NavItems: web.BuildNav("/"),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/firewall-requests", func(w http.ResponseWriter, r *http.Request) {
		if err := renderer.Render(w, "firewall_requests.gohtml", web.PageData{
			Title:    "04-042 防火牆申請",
			NavItems: web.BuildNav("/firewall-requests"),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/firewall-graph", func(w http.ResponseWriter, r *http.Request) {
		if err := renderer.Render(w, "firewall_graph.gohtml", web.PageData{
			Title:    "防火牆 3D 關聯圖",
			NavItems: web.BuildNav("/firewall-graph"),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/asset-inventory", func(w http.ResponseWriter, r *http.Request) {
		if err := renderer.Render(w, "asset_inventory.gohtml", web.PageData{
			Title:    "04-008 資訊資產清冊",
			NavItems: web.BuildNav("/asset-inventory"),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/application-change-requests", func(w http.ResponseWriter, r *http.Request) {
		if err := renderer.Render(w, "application_change_requests.gohtml", web.PageData{
			Title:    "04-052 應用系統功能需求更新建議",
			NavItems: web.BuildNav("/application-change-requests"),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/accounts", func(w http.ResponseWriter, r *http.Request) {
		if err := renderer.Render(w, "accounts.gohtml", web.PageData{
			Title:    "04-062 特殊權限帳號管理",
			NavItems: web.BuildNav("/accounts"),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/protection-baselines", func(w http.ResponseWriter, r *http.Request) {
		if err := renderer.Render(w, "protection_baselines.gohtml", web.PageData{
			Title:    "04-069 防護基準執行說明",
			NavItems: web.BuildNav("/protection-baselines"),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/platform-requests", func(w http.ResponseWriter, r *http.Request) {
		if err := renderer.Render(w, "platform_requests.gohtml", web.PageData{
			Title:    "04-078 系統平台申請",
			NavItems: web.BuildNav("/platform-requests"),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/firewall-requests.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/firewall-requests", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/firewall-graph.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/firewall-graph", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/asset-inventory.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/asset-inventory", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/application-change-requests.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/application-change-requests", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/accounts.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/accounts", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/protection-baselines.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/protection-baselines", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/platform-requests.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/platform-requests", http.StatusMovedPermanently)
	})

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
