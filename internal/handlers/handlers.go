package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"isms-privilege/internal/db"
	"isms-privilege/internal/mailer"
	"isms-privilege/internal/models"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	DB     *db.DB
	Mailer *mailer.Mailer
}

func New(d *db.DB, m *mailer.Mailer) *Handler {
	return &Handler{DB: d, Mailer: m}
}

// ---- helpers ----

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func idFromPath(r *http.Request, prefix string) (int, error) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, prefix), "/")
	if len(parts) == 0 || parts[0] == "" {
		return 0, fmt.Errorf("missing id")
	}
	return strconv.Atoi(parts[0])
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ---- Accounts CRUD ----

// GET /api/accounts?status=active&q=keyword
func (h *Handler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	q := r.URL.Query().Get("q")
	var (
		accounts []models.PrivilegedAccount
		err      error
	)
	if q != "" {
		accounts, err = h.DB.SearchAccounts(q)
	} else {
		accounts, err = h.DB.ListAccounts(status)
	}
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if accounts == nil {
		accounts = []models.PrivilegedAccount{}
	}
	writeJSON(w, 200, accounts)
}

// GET /api/accounts/{id}
func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/accounts/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	a, err := h.DB.GetAccount(id)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, 200, a)
}

// POST /api/accounts
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var a models.PrivilegedAccount
	if err := readJSON(r, &a); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	a.Creator = GetUserEmail(r)
	if a.Status == "" {
		a.Status = models.StatusActive
	}
	id, err := h.DB.CreateAccount(&a)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	a.ID = int(id)
	writeJSON(w, 201, a)
}

// PUT /api/accounts/{id}
func (h *Handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/accounts/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	var a models.PrivilegedAccount
	if err := readJSON(r, &a); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	a.ID = id
	existing, _ := h.DB.GetAccount(id)
	if existing != nil {
		a.Creator = existing.Creator
	}
	if err := h.DB.UpdateAccount(&a); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, a)
}

// DELETE /api/accounts/{id}
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/accounts/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.DB.DeleteAccount(id); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"message": "deleted"})
}

// ---- Email Notification ----

// POST /api/accounts/{id}/notify  → send single notification
func (h *Handler) NotifyAccount(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/accounts/"), "/notify")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	a, err := h.DB.GetAccount(id)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	if a.Email == "" {
		writeJSON(w, 400, map[string]string{"error": "no email address"})
		return
	}
	token, err := generateToken()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "token generation failed"})
		return
	}
	expiry := time.Now().AddDate(0, 0, 7)
	if err := h.DB.SetConfirmToken(id, token, expiry); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	logStatus := "sent"
	logMsg := "通知發送成功"
	if err := h.Mailer.SendConfirmEmail(a.Email, a.OwnerName, a.SystemName, a.IPAddress, a.AccountName, a.AccountType, token); err != nil {
		logStatus = "failed"
		logMsg = err.Error()
	}
	h.DB.AddNotificationLog(&models.NotificationLog{
		AccountID:   id,
		AccountName: a.AccountName,
		Email:       a.Email,
		Status:      logStatus,
		Message:     logMsg,
	})
	if logStatus == "failed" {
		writeJSON(w, 500, map[string]string{"error": logMsg, "debug": "token_set_ok"})
		return
	}
	writeJSON(w, 200, map[string]string{"message": "notification sent", "email": a.Email})
}

// POST /api/notify-all  → bulk notify all active accounts with email
func (h *Handler) NotifyAll(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.DB.GetActiveAccountsWithEmail()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	sent, failed := 0, 0
	for _, a := range accounts {
		token, err := generateToken()
		if err != nil {
			failed++
			continue
		}
		expiry := time.Now().AddDate(0, 0, 7)
		h.DB.SetConfirmToken(a.ID, token, expiry)
		logStatus := "sent"
		logMsg := "通知發送成功"
		if err := h.Mailer.SendConfirmEmail(a.Email, a.OwnerName, a.SystemName, a.IPAddress, a.AccountName, a.AccountType, token); err != nil {
			logStatus = "failed"
			logMsg = err.Error()
			failed++
		} else {
			sent++
		}
		h.DB.AddNotificationLog(&models.NotificationLog{
			AccountID:   a.ID,
			AccountName: a.AccountName,
			Email:       a.Email,
			Status:      logStatus,
			Message:     logMsg,
		})
	}
	writeJSON(w, 200, map[string]int{"sent": sent, "failed": failed, "total": len(accounts)})
}

// ---- Confirm Page (user-facing) ----

// GET /confirm?token=xxx&action=continue|stop
func (h *Handler) ConfirmAccount(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	action := r.URL.Query().Get("action")
	if token == "" || (action != "continue" && action != "stop") {
		http.Error(w, "無效的確認連結", 400)
		return
	}
	a, err := h.DB.GetAccountByToken(token)
	if err != nil {
		http.Error(w, "連結已失效或不存在，請聯繫管理員。", 404)
		return
	}
	if a.TokenExpiry != nil && a.TokenExpiry.Before(time.Now()) {
		http.Error(w, "確認連結已過期，請聯繫管理員重新發送。", 400)
		return
	}
	if err := h.DB.ConfirmAccount(a.ID, action); err != nil {
		http.Error(w, "系統錯誤，請稍後再試。", 500)
		return
	}
	actionText := "繼續使用"
	actionColor := "#28a745"
	if action == "stop" {
		actionText = "停用帳號"
		actionColor = "#dc3545"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="zh-TW"><head><meta charset="UTF-8"><title>確認完成</title>
<style>body{font-family:Arial,sans-serif;display:flex;justify-content:center;align-items:center;min-height:100vh;margin:0;background:#f4f6f8;}
.box{background:#fff;border-radius:12px;padding:40px;text-align:center;box-shadow:0 4px 20px rgba(0,0,0,.1);max-width:480px;width:90%%;}
.icon{font-size:60px;}.title{font-size:24px;font-weight:bold;color:%s;margin:16px 0 8px;}
.desc{color:#666;line-height:1.6;}.account{background:#f8f9fa;border-radius:8px;padding:12px;margin:16px 0;font-size:14px;}
a{display:inline-block;margin-top:20px;background:#1a3a5c;color:#fff;padding:10px 24px;border-radius:8px;text-decoration:none;}</style></head>
<body><div class="box">
<div class="icon">%s</div>
<div class="title">確認完成 — %s</div>
<p class="desc">您已成功完成特殊權限帳號盤點確認。</p>
<div class="account">系統：%s &nbsp;|&nbsp; 帳號：%s &nbsp;|&nbsp; 使用者：%s</div>
<p class="desc" style="font-size:13px;color:#888;">確認結果已記錄至 ISMS 系統，感謝您的配合。</p>
<a href="/">返回首頁</a></div></body></html>`,
		actionColor,
		map[string]string{"continue": "✅", "stop": "🔒"}[action],
		actionText,
		a.SystemName, a.AccountName, a.OwnerName)
}

// ---- Stats ----

// GET /api/stats
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.DB.Stats()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, stats)
}

// ---- Notification Logs ----

// GET /api/notification-logs
func (h *Handler) ListLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.DB.ListNotificationLogs()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if logs == nil {
		logs = []models.NotificationLog{}
	}
	writeJSON(w, 200, logs)
}

// RegisterRoutes 掛載所有路由到 mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// SPA & confirm page
	mux.HandleFunc("/confirm", h.ConfirmAccount)

	// Auth
	mux.HandleFunc("/api/login", h.Login)
	mux.HandleFunc("/api/logout", h.Logout)
	mux.HandleFunc("/oauth2callback", h.OAuth2Callback)
	mux.HandleFunc("/api/me", h.Me)

	// API
	mux.HandleFunc("/api/stats", AuthMiddleware(h.Stats))
	mux.HandleFunc("/api/notification-logs", AuthMiddleware(h.ListLogs))
	mux.HandleFunc("/api/notify-all", AuthMiddleware(h.NotifyAll))

	mux.HandleFunc("/api/accounts", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListAccounts(w, r)
		case http.MethodPost:
			h.CreateAccount(w, r)
		default:
			http.Error(w, "method not allowed", 405)
		}
	}))

	mux.HandleFunc("/api/accounts/", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/notify") {
			if r.Method == http.MethodPost {
				h.NotifyAccount(w, r)
			} else {
				http.Error(w, "method not allowed", 405)
			}
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.GetAccount(w, r)
		case http.MethodPut:
			h.UpdateAccount(w, r)
		case http.MethodDelete:
			h.DeleteAccount(w, r)
		default:
			http.Error(w, "method not allowed", 405)
		}
	}))
}
