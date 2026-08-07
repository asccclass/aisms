package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// key 型別避免 context key 衝突
type contextKey string
const userEmailKey contextKey = "user_email"

var oauth2Config *oauth2.Config

func InitOAuth2() {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URI")
	if clientID == "" || clientSecret == "" {
		fmt.Println("Warning: GOOGLE_CLIENT_ID or SECRET not set")
		return
	}
	oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func isDevMockLoginEnabled() bool {
	return os.Getenv("ENABLE_DEV_MOCK_LOGIN") == "true"
}

// 登入介面
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if oauth2Config == nil {
		http.Error(w, "OAuth not configured", 503)
		return
	}
	mockEmail := r.URL.Query().Get("mock_email")
	if mockEmail != "" && isDevMockLoginEnabled() {
		http.SetCookie(w, &http.Cookie{
			Name:     "admin_email",
			Value:    mockEmail,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "admin_name",
			Value:    "測試人員",
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: false, // allow JS to read
		})
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	url := oauth2Config.AuthCodeURL("state-token", oauth2.AccessTypeOnline, oauth2.SetAuthURLParam("prompt", "select_account"))
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Google callback
func (h *Handler) OAuth2Callback(w http.ResponseWriter, r *http.Request) {
	if oauth2Config == nil {
		http.Error(w, "OAuth not configured", 503)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		http.Redirect(w, r, "/?error=missing_code", http.StatusTemporaryRedirect)
		return
	}

	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		http.Redirect(w, r, "/?error=exchange_failed", http.StatusTemporaryRedirect)
		return
	}

	client := oauth2Config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil || resp.StatusCode != 200 {
		http.Redirect(w, r, "/?error=userinfo_failed", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	var ui struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	json.Unmarshal(data, &ui)

	if ui.Email == "" {
		http.Redirect(w, r, "/?error=no_email", http.StatusTemporaryRedirect)
		return
	}

	// Set simple cookie session
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_email",
		Value:    ui.Email,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "admin_name",
		Value:    ui.Name,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: false, // allow JS to read
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// 登出
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_email",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_name",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: false,
	})
	writeJSON(w, 200, map[string]string{"message": "登出成功"})
}

// 取得當前使用者
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("admin_email")
	if err != nil || c.Value == "" {
		writeJSON(w, 401, map[string]string{"error": "Unauthorized"})
		return
	}
	nameCookie, _ := r.Cookie("admin_name")
	name := ""
	if nameCookie != nil {
		name = nameCookie.Value
	}
	writeJSON(w, 200, map[string]string{
		"email": c.Value,
		"name":  name,
		"department": "系統管理部", // 預設或從資料庫帶出
	})
}

// AuthMiddleware 驗證登入狀態
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("admin_email")
		if err != nil || c.Value == "" {
			writeJSON(w, 401, map[string]string{"error": "Unauthorized"})
			return
		}
		// 將 email 放進 context 供後續使用
		ctx := context.WithValue(r.Context(), userEmailKey, c.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserEmail 從 context 取出
func GetUserEmail(r *http.Request) string {
	val := r.Context().Value(userEmailKey)
	if val != nil {
		return val.(string)
	}
	return ""
}
