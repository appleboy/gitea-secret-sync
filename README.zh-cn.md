# Gitea Secret Sync

[English](README.md) | [繁體中文](README.zh-tw.md) | [简体中文](README.zh-cn.md)

[![Go Report Card](https://goreportcard.com/badge/github.com/appleboy/gitea-secret-sync)](https://goreportcard.com/report/github.com/appleboy/gitea-secret-sync)
[![Lint and Testing](https://github.com/appleboy/gitea-secret-sync/actions/workflows/testing.yml/badge.svg)](https://github.com/appleboy/gitea-secret-sync/actions/workflows/testing.yml)
[![Trivy Security Scan](https://github.com/appleboy/gitea-secret-sync/actions/workflows/trivy.yml/badge.svg)](https://github.com/appleboy/gitea-secret-sync/actions/workflows/trivy.yml)

一个功能强大的 CLI 工具，用于在 Gitea 组织和仓库之间同步密钥。该工具可以批量更新多个 Gitea 组织和仓库的动作密钥，并支持预览模式和 SSL 验证选项。

## 功能特性

- 🔐 **批量密钥管理**：一次同步多个密钥到组织和仓库
- 🏢 **组织支持**：更新多个 Gitea 组织的密钥
- 📁 **仓库支持**：更新特定仓库的密钥（格式：`org/repo`）
- 🔍 **预览模式**：在不实际更新密钥的情况下预览更改
- 🔒 **SSL 配置**：为自托管的 Gitea 实例提供跳过 SSL 验证的选项
- 📁 **环境文件支持**：从 `.env` 文件加载配置
- 📊 **结构化日志**：具有结构化输出的清晰日志记录
- 🛡️ **优雅关闭**：优雅地处理中断信号

## 安装

### 从源代码安装

```bash
go install github.com/appleboy/gitea-secret-sync/cmd@latest
```

### 从发布版安装

从 [发布页面](https://github.com/appleboy/gitea-secret-sync/releases) 下载最新的二进制文件。

## 使用方法

### 命令行选项

```bash
gitea-secret-sync [选项]
```

**选项：**

- `-env-file string`：读取环境变量文件（默认：".env"）
- `-version`：显示版本信息并退出

### 环境变量

该工具支持两种环境变量格式：

1. 直接格式：`VARIABLE_NAME=value`
2. 输入格式：`INPUT_VARIABLE_NAME=value`（优先于直接格式）

#### 必需配置

| 变量           | 描述                                   | 示例                        |
| -------------- | -------------------------------------- | --------------------------- |
| `GITEA_SERVER` | Gitea 服务器 URL                       | `https://gitea.example.com` |
| `GITEA_TOKEN`  | Gitea 访问令牌                         | `your_gitea_token_here`     |
| `SECRETS`      | 要同步的密钥名称列表（逗号或换行分隔） | `SECRET1,SECRET2,SECRET3`   |

#### 可选配置

| 变量                | 描述                               | 默认值  | 示例                    |
| ------------------- | ---------------------------------- | ------- | ----------------------- |
| `GITEA_SKIP_VERIFY` | 跳过 SSL 证书验证                  | `false` | `true`                  |
| `ORGS`              | 要更新的组织列表（逗号或换行分隔） | -       | `org1,org2,org3`        |
| `REPOS`             | 要更新的仓库列表（逗号或换行分隔） | -       | `org1/repo1,org2/repo2` |
| `DRY_RUN`           | 启用预览模式（仅预览）             | `false` | `true`                  |

#### 密钥值

对于 `SECRETS` 中列出的每个密钥名称，您需要提供对应的环境变量：

```bash
# 如果 SECRETS=SECRET1,SECRET2,SECRET3
SECRET1=actual_secret_value_1
SECRET2=actual_secret_value_2
SECRET3=actual_secret_value_3
```

### 输入格式支持

该工具支持 `SECRETS`、`ORGS` 和 `REPOS` 变量的灵活输入格式：

#### 逗号分隔（传统格式）

```bash
SECRETS=SECRET1,SECRET2,SECRET3
ORGS=org1,org2,org3
REPOS=org1/repo1,org2/repo2,org3/repo3
```

#### 换行分隔格式

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

#### 混合格式（逗号与换行）

```bash
SECRETS="SECRET1,SECRET2
SECRET3"
ORGS="org1,org2
org3"
```

**注意：** 额外的空白字符和空项目会自动处理并过滤掉。

## 配置示例

### 使用 .env 文件

在项目目录中创建 `.env` 文件：

```bash
# Gitea 配置
GITEA_SERVER=https://gitea.example.com
GITEA_TOKEN=your_gitea_access_token
GITEA_SKIP_VERIFY=false

# 要同步的密钥（逗号分隔格式）
SECRETS=DATABASE_URL,API_KEY,JWT_SECRET

# 另一种选择：换行分隔格式
# SECRETS="DATABASE_URL
# API_KEY
# JWT_SECRET"

# 密钥值
DATABASE_URL=postgresql://user:pass@localhost/db
API_KEY=sk-1234567890abcdef
JWT_SECRET=super-secret-jwt

# 目标组织和仓库（逗号分隔）
ORGS=my-org,another-org
REPOS=my-org/repo1,my-org/repo2,another-org/repo3

# 另一种选择：换行分隔格式
# ORGS="my-org
# another-org"
# REPOS="my-org/repo1
# my-org/repo2
# another-org/repo3"

# 可选：启用预览模式
DRY_RUN=false
```

### 使用环境变量

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

### 使用 INPUT\_ 前缀（GitHub Actions 风格）

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

## 使用示例

### 基本使用

```bash
# 从 .env 文件加载配置并同步密钥
./gitea-secret-sync

# 使用自定义环境文件
./gitea-secret-sync -env-file production.env
```

### 预览模式

```bash
# 预览将要更改的内容而不实际更新
DRY_RUN=true ./gitea-secret-sync
```

### 仅更新组织

```bash
# 仅为特定组织更新密钥
export GITEA_SERVER=https://gitea.example.com
export GITEA_TOKEN=your_token
export SECRETS=SECRET1,SECRET2
export SECRET1=value1
export SECRET2=value2
export ORGS=org1,org2

./gitea-secret-sync
```

### 仅更新仓库

```bash
# 仅为特定仓库更新密钥
export GITEA_SERVER=https://gitea.example.com
export GITEA_TOKEN=your_token
export SECRETS=SECRET1,SECRET2
export SECRET1=value1
export SECRET2=value2
export REPOS=org1/repo1,org1/repo2,org2/repo3

./gitea-secret-sync
```

## 获取您的 Gitea 令牌

1. 登录您的 Gitea 实例
2. 前往 **设置** → **应用程序** → **生成新令牌**
3. 选择所需的作用域：
   - `write:organization`（用于组织密钥）
   - `write:repository`（用于仓库密钥）
4. 复制生成的令牌并将其用作 `GITEA_TOKEN`

## 错误处理

该工具包含完整的错误处理：

- **无效配置**：缺少必需变量将导致程序退出并显示错误消息
- **网络问题**：连接问题会记录详细的错误信息
- **API 错误**：Gitea API 错误会被捕获并记录相关上下文
- **无效仓库格式**：仓库名称必须采用 `org/repo` 格式
- **优雅关闭**：工具优雅地处理 SIGINT 和 SIGTERM 信号

## 日志记录

该工具使用结构化日志记录，具有以下日志级别：

- `INFO`：正常操作信息
- `WARN`：警告（例如预览模式通知）
- `ERROR`：阻止操作的错误条件

日志格式包括时间戳、级别、消息和相关上下文字段。

## 开发

### 从源代码构建

```bash
git clone https://github.com/appleboy/gitea-secret-sync.git
cd gitea-secret-sync
go build -o gitea-secret-sync ./cmd
```

### 运行测试

```bash
go test ./cmd
```

### 项目结构

```txt
cmd/
├── gitea.go           # 具有 SSL 配置的 Gitea 客户端包装器
├── main.go            # 主应用程序入口点和逻辑
├── util.go            # 环境变量和字符串解析的实用函数
├── util_test.go       # 实用函数的单元测试
└── util_split_test.go # 字符串分割功能的单元测试
```

### 主要组件

- **`gitea.go`**：实现具有 SSL 配置支持的 Gitea 客户端包装器
- **`main.go`**：包含主应用程序逻辑、信号处理和批处理
- **`util.go`**：提供读取环境变量的实用函数，支持 INPUT\_ 前缀和灵活的字符串解析（逗号/换行分隔）
- **`util_test.go`**：实用函数的单元测试
- **`util_split_test.go`**：新字符串分割功能的单元测试

## 贡献

1. Fork 此仓库
2. 创建您的功能分支（`git checkout -b feature/amazing-feature`）
3. 提交您的更改（`git commit -m 'Add some amazing feature'`）
4. 推送到分支（`git push origin feature/amazing-feature`）
5. 开启拉取请求

## 许可证

此项目采用 MIT 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件。

## 安全注意事项

- **令牌安全性**：安全地存储您的 Gitea 令牌，切勿将其提交到版本控制系统

- **SSL 验证**：仅在受信任的内部网络中禁用 SSL 验证（`GITEA_SKIP_VERIFY=true`）
- **预览模式**：在进行实际更改之前，请务必使用 `DRY_RUN=true` 进行测试
- **密钥值**：确保密钥值正确转义，且不包含可能导致解析问题的特殊字符

## 故障排除

### 常见问题

#### "missing gitea server or token"

- 确保已设置 `GITEA_SERVER` 和 `GITEA_TOKEN` 环境变量

#### "invalid repo format"

- 仓库名称必须采用 `org/repo` 格式（例如：`myorg/myrepo`）

#### "failed to update secrets"

- 检查您的 Gitea 令牌是否具有足够的权限
- 验证组织或仓库是否存在
- 确保可以访问 Gitea 服务器

#### SSL 证书错误

- 对于自签证书，请设置 `GITEA_SKIP_VERIFY=true`（不建议用于生产环境）
- 或正确配置 SSL 证书

### 寻求帮助

如果您遇到问题：

1. 启用预览模式以检查配置：`DRY_RUN=true`
2. 检查日志以查看详细的错误消息
3. 验证您的 Gitea 令牌权限
4. 测试与 Gitea 服务器的连接

## 相关项目

- [Gitea](https://gitea.io/) - 无痛的自托管 Git 服务
- [Gitea SDK](https://code.gitea.io/sdk) - Gitea API 的 Go SDK

---

用 ❤️ 为 Gitea 社区制作
