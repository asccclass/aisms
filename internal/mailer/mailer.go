package mailer

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"text/template"
	"time"
)

type Mailer struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
	BaseURL  string
}

// New 從環境變數建立 Mailer
func New() *Mailer {
	return &Mailer{
		Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		Port:     getEnv("SMTP_PORT", "587"),
		User:     getEnv("SMTP_USER", ""),
		Password: getEnv("SMTP_PASSWORD", ""),
		From:     getEnv("SMTP_FROM", "ISMS系統 <isms@example.com>"),
		BaseURL:  getEnv("SERVER_URL", "http://localhost:8080"),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

const confirmTmpl = `
<!DOCTYPE html>
<html lang="zh-TW">
<head><meta charset="UTF-8"><title>特殊權限帳號確認通知</title></head>
<body style="font-family:Arial,sans-serif;background:#f4f6f8;margin:0;padding:20px;">
<div style="max-width:600px;margin:0 auto;background:#fff;border-radius:10px;overflow:hidden;box-shadow:0 4px 12px rgba(0,0,0,.1);">
  <div style="background:linear-gradient(135deg,#1a3a5c,#2d6ca2);padding:30px;text-align:center;">
    <h1 style="color:#fff;margin:0;font-size:22px;">🔐 ISMS 特殊權限帳號確認通知</h1>
  </div>
  <div style="padding:30px;">
    <p>親愛的 <strong>{{.OwnerName}}</strong>，您好：</p>
    <p>您持有以下特殊權限帳號，依 ISMS 政策需定期盤點確認，請於 <strong>{{.Deadline}}</strong> 前完成線上確認。</p>
    <table style="width:100%;border-collapse:collapse;margin:20px 0;font-size:14px;">
      <tr style="background:#f0f4f8;"><td style="padding:10px;border:1px solid #ddd;font-weight:bold;">系統名稱</td><td style="padding:10px;border:1px solid #ddd;">{{.SystemName}}</td></tr>
      <tr><td style="padding:10px;border:1px solid #ddd;font-weight:bold;">IP 位址</td><td style="padding:10px;border:1px solid #ddd;">{{.IPAddress}}</td></tr>
      <tr style="background:#f0f4f8;"><td style="padding:10px;border:1px solid #ddd;font-weight:bold;">帳號名稱</td><td style="padding:10px;border:1px solid #ddd;">{{.AccountName}}</td></tr>
      <tr><td style="padding:10px;border:1px solid #ddd;font-weight:bold;">帳號種類</td><td style="padding:10px;border:1px solid #ddd;">{{.AccountType}}</td></tr>
    </table>
    <p>請點選下方按鈕確認是否繼續使用此帳號：</p>
    <div style="text-align:center;margin:30px 0;">
      <a href="{{.ContinueURL}}" style="background:#28a745;color:#fff;padding:14px 32px;border-radius:8px;text-decoration:none;font-size:16px;margin:0 10px;display:inline-block;">✅ 繼續使用</a>
      <a href="{{.StopURL}}" style="background:#dc3545;color:#fff;padding:14px 32px;border-radius:8px;text-decoration:none;font-size:16px;margin:0 10px;display:inline-block;">❌ 停用帳號</a>
    </div>
    <p style="font-size:12px;color:#888;">⚠️ 此連結有效期限為 7 天，逾期未確認將由管理員依規定處理。</p>
    <p style="font-size:12px;color:#888;">若您有任何問題，請聯繫資安部門。</p>
  </div>
  <div style="background:#f4f6f8;padding:15px;text-align:center;font-size:12px;color:#aaa;">
    ISMS 資訊安全管理系統 © {{.Year}}
  </div>
</div>
</body></html>
`

type ConfirmEmailData struct {
	OwnerName   string
	SystemName  string
	IPAddress   string
	AccountName string
	AccountType string
	ContinueURL string
	StopURL     string
	Deadline    string
	Year        int
}

// SendConfirmEmail 發送確認 Email
func (m *Mailer) SendConfirmEmail(to, ownerName, systemName, ipAddress, accountName, accountType, token string) error {
	data := ConfirmEmailData{
		OwnerName:   ownerName,
		SystemName:  systemName,
		IPAddress:   ipAddress,
		AccountName: accountName,
		AccountType: accountType,
		ContinueURL: fmt.Sprintf("%s/confirm?token=%s&action=continue", m.BaseURL, token),
		StopURL:     fmt.Sprintf("%s/confirm?token=%s&action=stop", m.BaseURL, token),
		Deadline:    time.Now().AddDate(0, 0, 7).Format("2006-01-02"),
		Year:        time.Now().Year(),
	}

	tmpl, err := template.New("confirm").Parse(confirmTmpl)
	if err != nil {
		return fmt.Errorf("template parse error: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("template execute error: %w", err)
	}

	msg := buildMIME(m.From, to, "【ISMS】特殊權限帳號確認通知 - "+accountName, buf.String())
	return m.sendSMTP(to, msg)
}

func (m *Mailer) sendSMTP(to string, msg []byte) error {
	addr := m.Host + ":" + m.Port
	auth := smtp.PlainAuth("", m.User, m.Password, m.Host)

	// Try TLS first (port 587 STARTTLS)
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial error: %w", err)
	}
	defer client.Close()

	tlsCfg := &tls.Config{ServerName: m.Host}
	if err := client.StartTLS(tlsCfg); err != nil {
		return fmt.Errorf("starttls error: %w", err)
	}
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth error: %w", err)
	}
	if err := client.Mail(m.User); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = w.Write(msg)
	return err
}

func buildMIME(from, to, subject, htmlBody string) []byte {
	var buf bytes.Buffer
	buf.WriteString("From: " + from + "\r\n")
	buf.WriteString("To: " + to + "\r\n")
	buf.WriteString("Subject: " + subject + "\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody)
	return buf.Bytes()
}
