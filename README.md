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

### 前端檔案結構

目前前端已將單頁應用的主要邏輯與 modal 拆分，避免所有功能都集中在 `index.html`：

```text
www/html/
├── index.html                  # 頁面骨架、主要區塊、script 掛載點
├── js/
│   ├── common.js               # 共用狀態、helper、登入檢查、頁面切換
│   ├── dashboard.js            # 儀表板資料載入與畫面渲染
│   ├── accounts.js             # 帳號 CRUD、通知、刪除、通知記錄
│   ├── platform-requests.js    # 04-078 系統平台申請 CRUD
│   ├── forms.js                # 表單管理、表單設定 modal 載入與儲存
│   ├── account-modals.js       # 帳號/通知 modal partial 的 lazy-load
│   └── export-docx.js          # DOCX 匯出流程、簽核欄記憶功能
└── partials/
    ├── account-modal.html      # 帳號新增/編輯 modal
    ├── notify-modal.html       # 發送通知確認 modal
    ├── export-docx-modal.html  # DOCX 匯出簽核欄 modal
    ├── form-config-modal.html  # 表單設定 modal
    └── platform-request-modal.html # 04-078 申請輸入 modal
```

說明：

- `index.html` 目前只負責主畫面骨架、共用 overlay 與各 feature 的掛載點。
- `js/` 目錄依功能切分腳本，方便逐步修改單一模組。
- `partials/` 目錄放可獨立維護的 modal HTML 片段，由前端在需要時動態載入。
- 若後續再新增功能，建議優先沿用這種做法，把新功能的 UI 與邏輯放到對應檔案，而不是回塞到 `index.html`。

### API 路由

| Method | Path | 說明 |
|--------|------|------|
| GET    | `/api/accounts` | 列出帳號（支援 `?status=` `?q=` 查詢）|
| POST   | `/api/accounts` | 新增帳號 |
| GET    | `/api/accounts/{id}` | 取得單一帳號 |
| PUT    | `/api/accounts/{id}` | 更新帳號 |
| DELETE | `/api/accounts/{id}` | 刪除帳號 |
| POST   | `/api/accounts/{id}/notify` | 發送 Email 確認通知 |
| GET    | `/api/accounts/export-docx` | 匯出目前帳號資料為 DOCX |
| GET    | `/api/platform-requests` | 列出 04-078 系統平台申請資料 |
| POST   | `/api/platform-requests` | 新增 04-078 系統平台申請資料 |
| GET    | `/api/platform-requests/{id}` | 取得單一 04-078 申請資料 |
| PUT    | `/api/platform-requests/{id}` | 更新單一 04-078 申請資料 |
| DELETE | `/api/platform-requests/{id}` | 刪除單一 04-078 申請資料 |
| GET    | `/api/platform-requests/{id}/export-docx` | 匯出單筆 04-078 申請資料為 DOCX |
| GET    | `/api/platform-requests/{id}/export-pdf` | 匯出單筆 04-078 申請資料為 PDF |
| POST   | `/api/notify-all` | 批次通知全部有效帳號 |
| GET    | `/api/stats` | 統計數字 |
| GET    | `/api/notification-logs` | 通知記錄 |
| GET    | `/confirm?token=&action=` | 使用者確認頁（繼續 / 停用）|

#### `platform-requests` 欄位用途

`ISMS-04-078 系統平台申請` 目前支援記錄下列申請端資訊：

- 填表日期、申請人姓名、單位、職稱、辦公室電話、Email、PI
- 應用系統名稱、英文簡稱、系統用途說明、預估使用人數
- 限院內使用、IP 位址限制、申請期間起訖
- 申請樣態、關機保留月數、關機/保留原因
- 環境類別、作業系統、其他 OS 理由、硬碟容量
- 特殊需求、網域名稱設定、其他需求
- 備份需求、特殊備份需求、不需備份理由
- 申請人簽章、主管簽章、狀態、備註

#### `GET /api/platform-requests/{id}/export-docx`

匯出單筆 **ISMS-04-078 系統平台申請表** 為 Word (`.docx`) 檔案。

說明：

- `id` 為申請資料主鍵。
- 匯出內容會帶入該筆申請的欄位資料，例如申請人、系統名稱、申請期間、環境、作業系統、備份需求與簽章欄位。
- 可額外帶入下列查詢參數，直接寫進「系統科辦理結果」區塊：
- `handler_name`：承辦人
- `manager_name`：權責主管
- `review_notes`：審查意見 / 安裝或移除路徑

範例：

```http
GET /api/platform-requests/12/export-docx?handler_name=王小明&manager_name=陳主管&review_notes=同意上架，請附測試報告
```

#### `GET /api/platform-requests/{id}/export-pdf`

匯出單筆 **ISMS-04-078 系統平台申請表** 為 PDF (`.pdf`) 檔案。

說明：

- `id` 為申請資料主鍵。
- 匯出內容會帶入該筆申請的欄位資料，例如申請人、系統名稱、申請期間、環境、作業系統、備份需求與簽章欄位。
- 可額外帶入下列查詢參數，直接寫進「系統科辦理結果」區塊：
- `handler_name`：承辦人
- `manager_name`：權責主管
- `review_notes`：審查意見 / 安裝或移除路徑
- 回傳內容型別為 `application/pdf`
- 此端點會先依原始 DOCX 樣板產生文件，再呼叫 LibreOffice (`soffice`) 轉成 PDF，以保留原樣板的表頭、表格與表尾版面。
- 若執行環境未安裝 LibreOffice，API 會回傳轉檔器不存在的錯誤訊息。

範例：

```http
GET /api/platform-requests/12/export-pdf?handler_name=王小明&manager_name=陳主管&review_notes=同意上架，請附測試報告
```

#### `GET /api/accounts/export-docx`

匯出 **ISMS-04-062 特殊權限帳號盤點清冊** 的 Word (`.docx`) 檔案，回傳內容型別為：

```http
application/vnd.openxmlformats-officedocument.wordprocessingml.document
```

支援查詢參數：

| 參數 | 說明 |
|------|------|
| `status` | 依帳號狀態篩選，例如 `active` / `pending` / `closed` / `expired` |
| `q` | 關鍵字搜尋，會套用和畫面列表相同的搜尋條件 |
| `inventory_by` | 匯出時寫入 DOCX 簽核欄的「盤點者」 |
| `group_leader` | 匯出時寫入 DOCX 簽核欄的「組長」 |

範例：

```http
GET /api/accounts/export-docx?status=active&inventory_by=王小明&group_leader=陳組長
```

說明：

- 若同時帶入 `q`，系統會以搜尋結果作為匯出內容。
- 若未帶 `q`，可用 `status` 匯出指定狀態的帳號資料。
- 若未帶 `inventory_by` / `group_leader`，系統會回退使用 `envfile` 中的 `DOCX_INVENTORY_BY` / `DOCX_GROUP_LEADER`。
- 前端畫面中的 `下載 DOCX` 按鈕，會先跳出輸入視窗，並將目前列表篩選條件一併帶入匯出。

---

## 快速啟動

### 1. 環境需求

- Go 1.21+（需含 CGO 支援 sqlite3）
- GCC / musl（Alpine 環境：`apk add gcc musl-dev sqlite-dev`）
- 若需使用 `GET /api/platform-requests/{id}/export-pdf`，執行環境必須額外安裝 LibreOffice，並確保 `soffice` 可由系統 `PATH` 呼叫

### 2. 設定 envfile

```ini
SystemName=ISMS資訊資產管理系統
PORT=8080
DB_PATH=data/isms.db
DocumentRoot=www/html
TemplateRoot=www/template
DOCUMENT_ROOT=www/html
SERVER_URL=http://your-server.example.com
OriginAllowList=http://localhost:8080
AllowMethods=GET;POST;PUT;DELETE;OPTIONS

# SMTP 設定（Gmail 範例，需開啟應用程式密碼）
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=abcd efgh ijkl mnop  # Gmail App Password
SMTP_FROM=ISMS系統 <your-email@gmail.com>

# DOCX 匯出簽核欄
DOCX_OWNER_DEPARTMENT=資安科
DOCX_INVENTORY_BY=
DOCX_GROUP_LEADER=

# 04-078 PDF 匯出字型
# 請保留 assets/fonts/NotoSansTC-VF.ttf，確保繁中 PDF 正常顯示

# Google 登入限制與單位推估
GOOGLE_ALLOWED_DOMAINS=example.org,sinica.edu.tw
GOOGLE_DOMAIN_DEPARTMENTS=example.org=資訊服務處;sinica.edu.tw=中央研究院
GOOGLE_WORKSPACE_ADMIN_EMAIL=admin@example.org
GOOGLE_WORKSPACE_SERVICE_ACCOUNT_FILE=/path/to/service-account.json
# 或直接放 JSON 內容
GOOGLE_WORKSPACE_SERVICE_ACCOUNT_JSON=
```

### Google Workspace 管理端設定步驟

若要讓系統自動取得登入者的單位、部門、職稱與組織資訊，程式會依序嘗試：

1. `Directory API`
2. `People API`
3. `hosted domain / email domain` fallback

其中前兩者需要由 Google Workspace 管理員先完成 service account 與 Domain-Wide Delegation 設定。

#### 1. 建立或選擇 Google Cloud 專案

1. 進入 Google Cloud Console。
2. 建立新專案，或選擇既有專案。
3. 確認這個專案就是你之後要放 OAuth Client 與 service account 的專案。

官方文件：

- Create access credentials: https://developers.google.com/workspace/guides/create-credentials

#### 2. 啟用需要的 API

在 Google Cloud Console：

1. 進入 `Menu > APIs & Services > Library`
2. 啟用下列 API：
3. `Admin SDK API`
4. `People API`

官方文件：

- Enable Google Workspace APIs: https://developers.google.com/workspace/guides/enable-apis
- Directory API Overview: https://developers.google.com/workspace/admin/directory/v1/guides

#### 3. 建立 Service Account

在 Google Cloud Console：

1. 進入 `Menu > IAM & Admin > Service Accounts`
2. 點 `Create Service Account`
3. 建立一個專用帳號，例如 `isms-workspace-reader`
4. 建立後進入該 service account 詳細頁
5. 啟用 `Domain-wide delegation`
6. 產生並下載 JSON 金鑰檔

你之後會需要：

- service account JSON 檔
- service account 的 `Client ID`，不是 OAuth Web Client ID

官方文件：

- Create service accounts: https://cloud.google.com/iam/docs/service-accounts-create

#### 4. 在 Google Admin Console 授權 Domain-Wide Delegation

你必須使用 **Google Workspace Super Admin** 帳號操作。

在 Google Admin Console：

1. 進入 `Menu > Security > Access and data control > API controls`
2. 找到 `Domain wide delegation`
3. 點 `Manage Domain Wide Delegation`
4. 點 `Add new`
5. 在 `Client ID` 填入剛剛 service account 的 numeric client ID
6. 在 `OAuth scopes` 填入這個專案需要的 scopes，逗號分隔：

```text
https://www.googleapis.com/auth/admin.directory.user.readonly,
https://www.googleapis.com/auth/user.organization.read
```

7. 點 `Authorize`

官方文件：

- Control API access with domain-wide delegation: https://knowledge.workspace.google.com/admin/apps/control-api-access-with-domain-wide-delegation
- Perform Google Workspace domain-wide delegation of authority: https://developers.google.com/workspace/cloud-search/docs/guides/delegation

補充：

- `admin.directory.user.readonly` 用來查 `Directory API users.get`
- `user.organization.read` 用來查 `People API people/me` 的 `organizations`
- 權限變更可能需要幾分鐘到一段時間才會生效

Directory / People API 欄位參考：

- Directory API `users.get`: https://developers.google.com/workspace/admin/directory/reference/rest/v1/users/get
- Directory API `User` resource: https://developers.google.com/workspace/admin/directory/reference/rest/v1/users
- People API `people.get`: https://developers.google.com/people/api/rest/v1/people/get
- People API `organizations` 欄位: https://developers.google.com/people/api/rest/v1/people

#### 5. 決定要用哪個管理員身分查 Directory API

本專案目前會用 `GOOGLE_WORKSPACE_ADMIN_EMAIL` 當 impersonation subject 去查 `Directory API`。

建議：

1. 使用一個可讀取使用者資料的管理員帳號
2. 至少要能讀取使用者基本資料與組織資訊
3. 若組織採最小權限原則，建議建立一個專用管理員帳號或自訂 Admin Role

如果沒有足夠權限，系統就會略過 Directory API，退回 People API 或網域 fallback。

#### 6. 在本專案填入環境變數

至少設定以下項目：

```ini
GOOGLE_ALLOWED_DOMAINS=example.org,sinica.edu.tw
GOOGLE_DOMAIN_DEPARTMENTS=example.org=資訊服務處;sinica.edu.tw=中央研究院
GOOGLE_WORKSPACE_ADMIN_EMAIL=admin@example.org
GOOGLE_WORKSPACE_SERVICE_ACCOUNT_FILE=/path/to/service-account.json
```

或改成直接放 JSON：

```ini
GOOGLE_WORKSPACE_SERVICE_ACCOUNT_JSON={...完整 service account JSON...}
```

說明：

- `GOOGLE_ALLOWED_DOMAINS`：限制哪些 Google Workspace 網域可登入
- `GOOGLE_DOMAIN_DEPARTMENTS`：當查不到正式部門欄位時，用網域做最後對照
- `GOOGLE_WORKSPACE_ADMIN_EMAIL`：Directory API impersonation 用的管理員帳號
- `GOOGLE_WORKSPACE_SERVICE_ACCOUNT_FILE` / `GOOGLE_WORKSPACE_SERVICE_ACCOUNT_JSON`：service account 憑證

#### 7. 本專案實際查詢順序

登入成功後，系統會這樣解析部門資訊：

1. 先查 `Directory API`
2. 在 Directory 內優先讀：
3. `customSchemas`
4. `organizations[].department / title / name`
5. `orgUnitPath`
6. 若 Directory 查不到，再查 `People API organizations`
7. 若還是沒有，最後才用 `hd` 或 `email domain` 搭配 `GOOGLE_DOMAIN_DEPARTMENTS`

因此，若你們希望資料最準確，建議優先：

1. 在 Google Workspace `customSchemas` 維護正式單位欄位
2. 或至少確保使用者 Profile 的 `organizations.department` 有填寫

#### 8. 驗證結果

設定完成後，可登入系統並呼叫：

```http
GET /api/me
```

預期會看到下列欄位之一或多個：

- `department`
- `department_source`
- `title`
- `organization_name`
- `org_unit_path`

其中 `department_source` 可能是：

- `directory`
- `directory_custom_schema`
- `people`
- `hd`

### 故障排除

#### `department_source` 為什麼只剩 `hd`

若 `/api/me` 回傳的 `department_source` 是 `hd`，表示系統前面兩層都沒有成功取到正式資料：

1. `Directory API` 沒有成功回傳可用欄位
2. `People API` 也沒有回傳 `organizations.department`
3. 最後只好退回 `hosted domain` 或 `email domain`

常見原因：

- `GOOGLE_WORKSPACE_SERVICE_ACCOUNT_FILE` / `GOOGLE_WORKSPACE_SERVICE_ACCOUNT_JSON` 沒有設定
- `GOOGLE_WORKSPACE_ADMIN_EMAIL` 沒有設定
- service account 沒有完成 DWD 授權
- `Admin SDK API` 或 `People API` 沒有啟用
- 該使用者在 Google Workspace 中根本沒有填 `organizations.department`
- 該使用者沒有對應的 `customSchemas` 欄位
- `orgUnitPath` 有值，但你期待的是更細的部門名稱

建議檢查順序：

1. 先看 `/api/me` 的 `department_source`
2. 再看 `org_unit_path`、`organization_name`、`title` 是否其實有值
3. 若這些都空白，優先檢查 DWD 與 API 啟用狀態

#### DWD 明明設了但還是查不到

若你已經完成 Domain-Wide Delegation，但仍然查不到部門資訊，常見問題如下：

1. **Client ID 填錯**

- 必須填 service account 的 numeric client ID
- 不能填 OAuth Web Client ID

2. **Scope 不完整**

本專案至少需要：

```text
https://www.googleapis.com/auth/admin.directory.user.readonly,
https://www.googleapis.com/auth/user.organization.read
```

少了任一個，就可能只剩部分資料可查，或直接 fallback。

3. **Admin impersonation subject 沒有權限**

- `GOOGLE_WORKSPACE_ADMIN_EMAIL` 必須是一個可讀取使用者資訊的管理員帳號
- 若這個帳號權限不足，Directory API 可能會失敗

4. **API 尚未啟用**

- `Admin SDK API` 沒啟用時，Directory API 一定查不到
- `People API` 沒啟用時，People fallback 也不會成功

5. **授權尚未傳播完成**

- DWD 或 API 權限更新後，可能需要一段時間才完全生效
- 建議稍等幾分鐘後再重新登入測試

6. **資料其實不存在**

- DWD 成功不代表使用者一定有填部門欄位
- 如果 Directory 的 `customSchemas`、`organizations` 和 People API 的 `organizations` 都沒有資料，最後仍會是 `hd`

#### `customSchemas` / `organizations` / `orgUnitPath` 各自適合放什麼資訊

這三者都能拿來做部門判讀，但適合的用途不同：

1. `customSchemas`

最適合：

- 你們自己定義的正式欄位
- 例如：`department_code`、`department_name`、`division`、`cost_center`
- 需要穩定欄位名、方便系統整合與稽核時

優點：

- 最可控
- 最適合企業內部系統整合
- 不會因個人 profile 填寫習慣不同而亂掉

建議：

- 若你們有正式組織欄位需求，優先使用這個

2. `organizations`

最適合：

- 放個人工作資訊
- 例如：部門、職稱、組織名稱

優點：

- 結構完整，包含 `department`、`title`、`name`
- People API 和 Directory API 都可能讀得到類似資訊

限制：

- 有時資料依賴管理端或使用者是否有維護
- 一致性不一定比 `customSchemas` 好

建議：

- 若沒有自訂 schema，可先用這個
- 尤其 `department + title + organization_name` 很適合顯示在使用者資訊頁

3. `orgUnitPath`

最適合：

- 反映 Google Workspace 管理上的 OU 結構
- 例如：`/corp/security/infra`

優點：

- 常常比 `organizations.department` 更穩定
- 很適合做權限、政策、群組或租戶管理邏輯

限制：

- OU 不一定等於人事上的正式部門名稱
- 有些公司 OU 是為了設備政策、帳號管理，不是為了 HR 組織

建議：

- 適合當 fallback 或輔助判斷
- 不建議直接假設 OU 就等於對外顯示的部門名稱，除非你們內部本來就是這樣設計

#### 實務建議

若你們要的是「系統中穩定顯示正式單位」，建議優先順序如下：

1. `customSchemas`
2. `organizations.department`
3. `orgUnitPath`
4. `hd / email domain`

這也就是本專案目前的設計方向：先找正式欄位，再逐層 fallback，而不是一開始就用網域硬猜。

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

補充：

- 目前 Dockerfile 已包含 LibreOffice，因此容器內可直接使用 `04-078` 的 PDF 匯出功能。
- 若是在本機 Windows / Linux 直接執行 binary，請先自行安裝 LibreOffice，並確認 `soffice` 或 `soffice.exe` 可在命令列執行。

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
