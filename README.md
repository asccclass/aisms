# ISMS資訊資產管理系統

依據文件 **ISMS-04-062** 特殊權限帳號盤點清冊規格，以 Go 語言 + [sherryserver](https://github.com/asccclass/sherryserver) 框架建置之完整 CRUD 管理系統，含 Email 通知使用者線上確認功能。

---

## 功能一覽

| 功能 | 說明 |
|------|------|
| **帳號 CRUD** | 新增 / 查詢（搜尋、篩選） / 修改 / 刪除特殊權限帳號 |
| **Email 通知** | 對單一帳號或全體有效帳號批次發送確認通知 |
| **線上確認** | 使用者點擊 Email 連結直接線上確認「繼續使用」或「停用帳號」|
| **狀態追蹤** | active / pending（待確認）/ closed / expired 四種狀態 |
| **通知記錄** | 完整 Email 發送紀錄（成功 / 失敗 / 時間）|
| **儀表板** | 統計卡片 + 最近帳號清單 |

---

## 系統架構

```
isms-privilege/
├── cmd/
│   └── main.go              # 主程式（sherryserver 入口）
├── internal/
│   ├── models/
│   │   └── account.go       # 資料模型
│   ├── db/
│   │   └── db.go            # SQLite CRUD 層
│   ├── mailer/
│   │   └── mailer.go        # SMTP Email 服務
│   └── handlers/
│       └── handlers.go      # HTTP API 路由
├── www/html/
│   └── index.html           # 前端單頁應用 (SPA)
├── data/                    # SQLite 資料庫（自動建立）
├── envfile                  # 環境設定
├── Makefile
└── Dockerfile
```

### API 路由

| Method | Path | 說明 |
|--------|------|------|
| GET    | `/api/accounts` | 列出帳號（支援 `?status=` `?q=` 查詢）|
| POST   | `/api/accounts` | 新增帳號 |
| GET    | `/api/accounts/{id}` | 取得單一帳號 |
| PUT    | `/api/accounts/{id}` | 更新帳號 |
| DELETE | `/api/accounts/{id}` | 刪除帳號 |
| POST   | `/api/accounts/{id}/notify` | 發送 Email 確認通知 |
| POST   | `/api/notify-all` | 批次通知全部有效帳號 |
| GET    | `/api/stats` | 統計數字 |
| GET    | `/api/notification-logs` | 通知記錄 |
| GET    | `/confirm?token=&action=` | 使用者確認頁（繼續 / 停用）|

---

## 快速啟動

### 1. 環境需求

- Go 1.21+（需含 CGO 支援 sqlite3）
- GCC / musl（Alpine 環境：`apk add gcc musl-dev sqlite-dev`）

### 2. 設定 envfile

```ini
PORT=8080
DB_PATH=data/isms.db
DOCUMENT_ROOT=www/html
SERVER_URL=http://your-server.example.com

# SMTP 設定（Gmail 範例，需開啟應用程式密碼）
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=abcd efgh ijkl mnop  # Gmail App Password
SMTP_FROM=ISMS系統 <your-email@gmail.com>
```

### 3. 執行

```bash
# 初始化目錄
make init

# 下載依賴
go mod tidy

# 直接執行
make run

# 或編譯後執行
make build && ./isms-server
```

### 4. Docker 執行

```bash
docker build -t isms-privilege .
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  --env-file envfile \
  --name isms \
  isms-privilege
```

---

## 確認信流程

```
管理員 → [發送通知] → Email 送達使用者
                           ↓
              使用者點擊 [繼續使用] 或 [停用帳號]
                           ↓
              系統更新帳號狀態 + 記錄確認時間
```

確認連結格式：
```
http://your-server.com/confirm?token=<32bytes_hex>&action=continue
http://your-server.com/confirm?token=<32bytes_hex>&action=stop
```

- Token 有效期：**7 天**
- 超過有效期需管理員重新發送通知

---

## 資料庫結構

### `privileged_accounts`
| 欄位 | 說明 |
|------|------|
| id | 主鍵 |
| system_name | 系統名稱 |
| ip_address | IP 位址 |
| inventory_date | 盤點日期 (YYYYMMDD) |
| account_name | 帳號名稱 |
| account_type | 帳號種類（預設/系統管理用/關閉帳號）|
| department | 單位 |
| owner_name | 姓名 |
| email | Email（用於確認通知）|
| passphrase_rotate | 是否定期變更（是/否/NA）|
| status | 狀態（active/pending/closed/expired）|
| confirm_token | 確認 Token |
| token_expiry | Token 過期時間 |
| last_confirmed_at | 最後確認時間 |

### `notification_logs`
發送記錄（account_id, email, sent_at, status, message）

---

## SMTP 設定建議

### Gmail App Password
1. 啟用二步驟驗證
2. 前往 [Google App Passwords](https://myaccount.google.com/apppasswords)
3. 產生應用程式密碼，填入 `SMTP_PASSWORD`

### 企業 Exchange / Office365
```ini
SMTP_HOST=smtp.office365.com
SMTP_PORT=587
```
