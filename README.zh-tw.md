# Gitea Secret Sync

[English](README.md) | [繁體中文](README.zh-tw.md) | [简体中文](README.zh-cn.md)

[![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/gitea-secret-sync)](https://goreportcard.com/report/github.com/appleboy/gitea-secret-sync)
[![Lint and Testing](https://github.com/appleboy/gitea-secret-sync/actions/workflows/testing.yml/badge.svg)](https://github.com/appleboy/gitea-secret-sync/actions/workflows/testing.yml)

一個功能強大的 CLI 工具，用於在 Gitea 組織和倉庫之間同步機密。此工具可以批量更新多個 Gitea 組織和倉庫的動作機密，並支援預覽模式和 SSL 驗證選項。

## 功能特色

- 🔐 **批量機密管理**：一次同步多個機密到組織和倉庫
- 🏢 **組織支援**：更新多個 Gitea 組織的機密
- 📁 **倉庫支援**：更新特定倉庫的機密（格式：`org/repo`）
- 🔍 **預覽模式**：在不實際更新機密的情況下預覽變更
- 🔒 **SSL 配置**：為自主託管的 Gitea 實例提供跳過 SSL 驗證的選項
- 📁 **環境檔案支援**：從 `.env` 檔案載入配置
- 📊 **結構化日誌**：具有結構化輸出的清晰日誌記錄
- 🛡️ **優雅關閉**：優雅地處理中斷信號

## 安裝

### 從原始碼安裝

```bash
go install github.com/appleboy/gitea-secret-sync/cmd@latest
```

### 從發行版安裝

從 [發行頁面](https://github.com/appleboy/gitea-secret-sync/releases) 下載最新的二進位檔案。

## 使用方法

### 命令列選項

```bash
gitea-secret-sync [選項]
```

**選項：**

- `-env-file string`：讀取環境變數檔案（預設：".env"）
- `-version`：顯示版本資訊並退出

### 環境變數

此工具支援兩種環境變數格式：

1. 直接格式：`VARIABLE_NAME=value`
2. 輸入格式：`INPUT_VARIABLE_NAME=value`（優先於直接格式）

#### 必需配置

| 變數           | 描述                               | 範例                        |
| -------------- | ---------------------------------- | --------------------------- |
| `GITEA_SERVER` | Gitea 伺服器 URL                   | `https://gitea.example.com` |
| `GITEA_TOKEN`  | Gitea 存取權杖                     | `your_gitea_token_here`     |
| `SECRETS`      | 要同步的機密名稱清單（逗號或換行分隔） | `SECRET1,SECRET2,SECRET3`   |

#### 可選配置

| 變數                | 描述                           | 預設值  | 範例                    |
| ------------------- | ------------------------------ | ------- | ----------------------- |
| `GITEA_SKIP_VERIFY` | 跳過 SSL 憑證驗證              | `false` | `true`                  |
| `ORGS`              | 要更新的組織清單（逗號或換行分隔） | -       | `org1,org2,org3`        |
| `REPOS`             | 要更新的倉庫清單（逗號或換行分隔） | -       | `org1/repo1,org2/repo2` |
| `DRY_RUN`           | 啟用預覽模式（僅預覽）         | `false` | `true`                  |

#### 機密值

對於 `SECRETS` 中列出的每個機密名稱，您需要提供對應的環境變數：

```bash
# 如果 SECRETS=SECRET1,SECRET2,SECRET3
SECRET1=actual_secret_value_1
SECRET2=actual_secret_value_2
SECRET3=actual_secret_value_3
```

### 輸入格式支援

此工具支援 `SECRETS`、`ORGS` 和 `REPOS` 變數的靈活輸入格式：

#### 逗號分隔（傳統格式）

```bash
SECRETS=SECRET1,SECRET2,SECRET3
ORGS=org1,org2,org3
REPOS=org1/repo1,org2/repo2,org3/repo3
```

#### 換行分隔格式

```bash
SECRETS="SECRET1
SECRET2
SECRET3"

ORGS="org1
org2
org3"

REPOS="org1/repo1
org2/repo2
org3/repo3"
```

#### 混合格式（逗號與換行）

```bash
SECRETS="SECRET1,SECRET2
SECRET3"
ORGS="org1,org2
org3"
```

**注意：** 額外的空白字符和空項目會自動處理並過濾掉。

## 配置範例

### 使用 .env 檔案

在專案目錄中建立 `.env` 檔案：

```bash
# Gitea 配置
GITEA_SERVER=https://gitea.example.com
GITEA_TOKEN=your_gitea_access_token
GITEA_SKIP_VERIFY=false

# 要同步的機密（逗號分隔格式）
SECRETS=DATABASE_URL,API_KEY,JWT_SECRET

# 另一種選擇：換行分隔格式
# SECRETS="DATABASE_URL
# API_KEY
# JWT_SECRET"

# 機密值
DATABASE_URL=postgresql://user:pass@localhost/db
API_KEY=sk-1234567890abcdef
JWT_SECRET=super-secret-jwt

# 目標組織和倉庫（逗號分隔）
ORGS=my-org,another-org
REPOS=my-org/repo1,my-org/repo2,another-org/repo3

# 另一種選擇：換行分隔格式
# ORGS="my-org
# another-org"
# REPOS="my-org/repo1
# my-org/repo2
# another-org/repo3"

# 可選：啟用預覽模式
DRY_RUN=false
```

### 使用環境變數

```bash
export GITEA_SERVER=https://gitea.example.com
export GITEA_TOKEN=your_gitea_access_token
export SECRETS=DATABASE_URL,API_KEY,JWT_SECRET
export DATABASE_URL=postgresql://user:pass@localhost/db
export API_KEY=sk-1234567890abcdef
export JWT_SECRET=super-secret-jwt
export ORGS=my-org,another-org
export REPOS=my-org/repo1,my-org/repo2
```

### 使用 INPUT\_ 前綴（GitHub Actions 樣式）

```bash
export INPUT_GITEA_SERVER=https://gitea.example.com
export INPUT_GITEA_TOKEN=your_gitea_access_token
export INPUT_SECRETS=DATABASE_URL,API_KEY,JWT_SECRET
export INPUT_DATABASE_URL=postgresql://user:pass@localhost/db
export INPUT_API_KEY=sk-1234567890abcdef
export INPUT_JWT_SECRET=super-secret-jwt
export INPUT_ORGS=my-org,another-org
export INPUT_REPOS=my-org/repo1,my-org/repo2
```

## 使用範例

### 基本使用

```bash
# 從 .env 檔案載入配置並同步機密
./gitea-secret-sync

# 使用自訂環境檔案
./gitea-secret-sync -env-file production.env
```

### 預覽模式

```bash
# 預覽將要變更的內容而不實際更新
DRY_RUN=true ./gitea-secret-sync
```

### 僅更新組織

```bash
# 僅為特定組織更新機密
export GITEA_SERVER=https://gitea.example.com
export GITEA_TOKEN=your_token
export SECRETS=SECRET1,SECRET2
export SECRET1=value1
export SECRET2=value2
export ORGS=org1,org2

./gitea-secret-sync
```

### 僅更新倉庫

```bash
# 僅為特定倉庫更新機密
export GITEA_SERVER=https://gitea.example.com
export GITEA_TOKEN=your_token
export SECRETS=SECRET1,SECRET2
export SECRET1=value1
export SECRET2=value2
export REPOS=org1/repo1,org1/repo2,org2/repo3

./gitea-secret-sync
```

## 取得您的 Gitea 權杖

1. 登入您的 Gitea 實例
2. 前往 **設定** → **應用程式** → **產生新權杖**
3. 選擇所需的範圍：
   - `write:organization`（用於組織機密）
   - `write:repository`（用於倉庫機密）
4. 複製產生的權杖並將其用作 `GITEA_TOKEN`

## 錯誤處理

此工具包含完整的錯誤處理：

- **無效配置**：缺少必需變數將導致程式退出並顯示錯誤訊息
- **網路問題**：連線問題會記錄詳細的錯誤資訊
- **API 錯誤**：Gitea API 錯誤會被捕獲並記錄相關內容
- **無效倉庫格式**：倉庫名稱必須採用 `org/repo` 格式
- **優雅關閉**：工具優雅地處理 SIGINT 和 SIGTERM 信號

## 日誌記錄

此工具使用結構化日誌記錄，具有以下日誌級別：

- `INFO`：正常操作資訊
- `WARN`：警告（例如預覽模式通知）
- `ERROR`：阻止操作的錯誤條件

日誌格式包括時間戳、級別、訊息和相關內容欄位。

## 開發

### 從原始碼建置

```bash
git clone https://github.com/appleboy/gitea-secret-sync.git
cd gitea-secret-sync
go build -o gitea-secret-sync ./cmd
```

### 執行測試

```bash
go test ./cmd
```

### 專案結構

```txt
cmd/
├── gitea.go           # 具有 SSL 配置的 Gitea 客戶端包裝器
├── main.go            # 主應用程式入口點和邏輯
├── util.go            # 環境變數和字串解析的實用函數
├── util_test.go       # 實用函數的單元測試
└── util_split_test.go # 字串分割功能的單元測試
```

### 主要元件

- **`gitea.go`**：實現具有 SSL 配置支援的 Gitea 客戶端包裝器
- **`main.go`**：包含主應用程式邏輯、信號處理和批處理
- **`util.go`**：提供讀取環境變數的實用函數，支援 INPUT\_ 前綴和靈活的字串解析（逗號/換行分隔）
- **`util_test.go`**：實用函數的單元測試
- **`util_split_test.go`**：新字串分割功能的單元測試
- **`util_split_test.go`**：新字串分割功能的單元測試

## 貢獻

1. Fork 此倉庫
2. 建立您的功能分支（`git checkout -b feature/amazing-feature`）
3. 提交您的變更（`git commit -m 'Add some amazing feature'`）
4. 推送到分支（`git push origin feature/amazing-feature`）
5. 開啟拉取請求

## 授權條款

此專案採用 MIT 授權條款 - 詳情請參閱 [LICENSE](LICENSE) 檔案。

## 安全注意事項

- **權杖安全性**：安全地儲存您的 Gitea 權杖，切勿將其提交到版本控制系統
- **SSL 驗證**：僅在受信任的內部網路中停用 SSL 驗證（`GITEA_SKIP_VERIFY=true`）
- **預覽模式**：在進行實際變更之前，請務必使用 `DRY_RUN=true` 進行測試
- **機密值**：確保機密值正確轉義，且不包含可能導致解析問題的特殊字符

## 故障排除

### 常見問題

#### "missing gitea server or token"

- 確保已設定 `GITEA_SERVER` 和 `GITEA_TOKEN` 環境變數

#### "invalid repo format"

- 倉庫名稱必須採用 `org/repo` 格式（例如：`myorg/myrepo`）

#### "failed to update secrets"

- 檢查您的 Gitea 權杖是否具有足夠的權限
- 驗證組織或倉庫是否存在
- 確保可以存取 Gitea 伺服器

#### SSL 憑證錯誤

- 對於自簽憑證，請設定 `GITEA_SKIP_VERIFY=true`（不建議用於生產環境）
- 或正確配置 SSL 憑證

### 尋求協助

如果您遇到問題：

1. 啟用預覽模式以檢查配置：`DRY_RUN=true`
2. 檢查日誌以查看詳細的錯誤訊息
3. 驗證您的 Gitea 權杖權限
4. 測試與 Gitea 伺服器的連線

## 相關專案

- [Gitea](https://gitea.io/) - 無痛的自主託管 Git 服務
- [Gitea SDK](https://code.gitea.io/sdk) - Gitea API 的 Go SDK

---

用 ❤️ 為 Gitea 社群製作
