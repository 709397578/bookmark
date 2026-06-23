# Bookmark Backend (Go)

Bookmark 书签管理系统的后端 API 服务，使用 Go 语言编写。前端代码位于[仓库根目录](../README.md)。

## 🛠️ 技术栈

- **框架**: [Gin](https://github.com/gin-gonic/gin) Web Framework
- **数据库**: PostgreSQL 或 SQLite（由 `DB_DRIVER` 切换）+ GORM
- **认证**: JWT (`golang-jwt/jwt/v5`)
- **密码加密**: bcrypt (`golang.org/x/crypto`)
- **网页快照**: chromedp（可选，需安装 Chromium）

> Go 版本：`go 1.25`（见 `go.mod`）

## 📁 项目结构

```
backend/
├── cmd/              # 主程序入口
│   └── main.go
├── config/           # 配置加载（读取 .env / 环境变量）
│   └── config.go
├── handlers/         # API 处理器
│   ├── auth_handler.go
│   ├── bookmark_handler.go
│   ├── collection_handler.go
│   ├── folder_handler.go
│   └── setting_handler.go
├── internal/         # 内部包
│   └── database.go   # 数据库连接与 AutoMigrate
├── middleware/       # 中间件
│   ├── auth.go          # JWT 认证 + 管理员校验 + CORS
│   └── optional_auth.go # 可选认证（公开接口附带用户信息）
├── models/           # 数据模型
│   └── models.go
├── pkg/              # 可复用包
│   └── snapshot/     # 网页快照生成
├── routes/           # 路由配置
│   └── routes.go
├── utils/            # 工具函数
│   ├── jwt.go
│   ├── password.go
│   └── response.go
├── uploads/          # 上传文件目录（快照 / 头像）
├── .env              # 环境变量配置（不入库）
├── .env.example      # 环境变量示例
├── Dockerfile        # 多阶段构建
├── go.mod / go.sum   # Go 模块依赖
└── README.md
```

## 🚀 安装和运行

### 1. 安装依赖

```bash
cd backend
go mod download
```

### 2. 配置环境变量

复制 `.env.example` 为 `.env` 并按需修改：

```bash
cp .env.example .env
```

> ⚠️ **重要**：`JWT_SECRET` 不能为空，也不能保留默认值 `your-secret-key-change-in-production`，否则程序启动时会直接 `log.Fatal` 退出（见 `config/config.go`）。

### 3. 选择数据库驱动

在 `.env` 中设置 `DB_DRIVER`：

- `postgres`（默认）：使用 `DATABASE_URL` 连接 PostgreSQL
- `sqlite`：使用 `DB_PATH` 指定的 SQLite 文件，构建时需 `CGO_ENABLED=1`

### 4. 运行服务器

```bash
# 开发模式
go run cmd/main.go

# 或者构建后运行
go build -o pintree-backend cmd/main.go
./pintree-backend
```

服务默认监听 `http://localhost:8080`，可通过 `PORT` 修改。

首次启动时若不存在管理员账户，会按 `ADMIN_EMAIL` / `ADMIN_PASSWORD` 自动创建默认管理员。

## 🔧 环境变量

| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| `PORT` | 服务端口 | `8080` |
| `DB_DRIVER` | 数据库驱动：`postgres` / `sqlite` | `postgres` |
| `DATABASE_URL` | PostgreSQL DSN（`DB_DRIVER=postgres` 时生效） | — |
| `DB_PATH` | SQLite 文件路径（`DB_DRIVER=sqlite` 时生效） | `./data/pintree.db` |
| `JWT_SECRET` | JWT 签名密钥（**必填，必须修改默认值**） | — |
| `JWT_EXPIRATION` | Token 有效期（秒） | `86400`（24 小时） |
| `FRONTEND_URL` | 前端地址（CORS 白名单） | `http://localhost:5173` |
| `UPLOAD_PATH` | 上传文件根目录 | `./uploads` |
| `SNAPSHOT_DIR` | 网页快照目录 | `./uploads/snapshots` |
| `CHROME_PATH` | Chromium 可执行文件路径（快照功能） | — |
| `ADMIN_EMAIL` | 默认管理员邮箱（无管理员时自动创建） | `admin@pintree.com` |
| `ADMIN_PASSWORD` | 默认管理员密码 | `admin123` |
| `ADMIN_NAME` | 默认管理员昵称 | `管理员` |

## 📡 API 端点

所有接口统一前缀 `/api`。响应统一为 `utils.Response` 格式（见文末）。

### 健康检查 & 静态资源

- `GET /health` — 健康检查
- `GET /snapshots/*` — 快照文件静态服务
- `GET /avatars/*` — 头像文件静态服务

### 认证 API（`/api/auth`）

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| POST | `/auth/register` | 公开 | 用户注册 |
| POST | `/auth/login` | 公开 | 用户登录 |
| GET | `/auth/profile` | 需认证 | 获取当前用户信息 |
| POST | `/auth/avatar` | 需认证 | 上传头像 |
| DELETE | `/auth/avatar` | 需认证 | 删除头像 |
| PUT | `/auth/password` | 需认证 | 修改密码 |
| GET | `/auth/users` | 管理员 | 用户列表 |
| PUT | `/auth/users/:id` | 管理员 | 更新用户 |
| DELETE | `/auth/users/:id` | 管理员 | 删除用户 |

### 收藏夹 API（`/api/collections`）

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/collections` | 公开（可选认证） | 收藏夹列表，支持 `publicOnly` |
| GET | `/collections/:id` | 公开（可选认证） | 获取单个收藏夹 |
| GET | `/collections/slug/:slug` | 公开（可选认证） | 根据 slug 获取 |
| POST | `/collections` | 需认证 | 创建收藏夹 |
| PUT | `/collections/:id` | 需认证 | 更新收藏夹 |
| DELETE | `/collections/:id` | 需认证 | 删除收藏夹 |
| PUT | `/collections/batch/sort` | 需认证 | 批量更新排序 |

### 书签 API（`/api/bookmarks`）

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/bookmarks` | 公开（可选认证） | 书签列表，支持分页 / `collectionId` / `folderId` |
| GET | `/bookmarks/search?q=` | 公开（可选认证） | 搜索书签 |
| GET | `/bookmarks/:id` | 公开（可选认证） | 获取单个书签 |
| POST | `/bookmarks` | 需认证 | 创建书签 |
| PUT | `/bookmarks/:id` | 需认证 | 更新书签 |
| DELETE | `/bookmarks/:id` | 需认证 | 删除书签 |
| POST | `/bookmarks/:id/snapshot` | 需认证 | 生成网页快照 |
| GET | `/bookmarks/export` | 需认证 | 导出书签（JSON） |
| GET | `/bookmarks/export/html` | 需认证 | 导出书签（浏览器 HTML） |
| POST | `/bookmarks/import/html` | 需认证 | 导入浏览器书签 HTML |
| POST | `/bookmarks/batch/delete` | 需认证 | 批量删除 |
| PUT | `/bookmarks/batch/move` | 需认证 | 批量移动 |
| PUT | `/bookmarks/batch/sort` | 需认证 | 批量更新排序 |

### 文件夹 API（`/api/folders`）

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/folders` | 公开（可选认证） | 文件夹列表，支持 `collectionId` |
| GET | `/folders/:id` | 公开（可选认证） | 获取单个文件夹 |
| POST | `/folders` | 需认证 | 创建文件夹 |
| PUT | `/folders/:id` | 需认证 | 更新文件夹 |
| DELETE | `/folders/:id` | 需认证 | 删除文件夹 |
| PUT | `/folders/batch/sort` | 需认证 | 批量更新排序 |

### 设置 API（`/api/settings`）

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/settings` | 公开 | 获取所有设置 |
| GET | `/settings/public` | 公开 | 获取公开设置 |
| GET | `/settings/initSettings` | 公开 | 初始化默认设置 |
| GET | `/settings/:key` | 公开 | 获取单个设置 |
| PUT | `/settings/:key` | 管理员 | 更新设置 |

### 其他

- `GET /api/search/bookmarks` — 书签搜索（与 `/api/bookmarks/search` 等价）
- `GET /api/tags` — 标签接口（占位，返回空数组）

## 🔐 认证方式

需要认证的接口须在请求头携带 JWT：

```
Authorization: Bearer <your-jwt-token>
```

公开接口使用「可选认证」中间件：未登录可访问，登录后会在上下文中携带用户信息。管理员接口在认证基础上额外校验用户角色。

## 🗄️ 数据库迁移

首次运行时由 GORM `AutoMigrate` 自动建表，无需手动迁移。

## 🔁 跨域配置

通过 `middleware.CORS` 处理，允许来源由 `FRONTEND_URL` 指定。

## 📦 统一响应格式

成功响应：

```json
{
  "success": true,
  "message": "操作成功",
  "data": { }
}
```

错误响应：

```json
{
  "success": false,
  "error": "错误信息"
}
```

## 🐳 Docker 部署

`backend/Dockerfile` 使用多阶段构建，支持以下构建参数：

| Build Arg | 说明 |
| --- | --- |
| `CGO_ENABLED` | `1` 时启用 SQLite（需 gcc），`0` 时仅 PostgreSQL |
| `INCLUDE_CHROME` | `1` 时安装 Chromium（用于网页快照，约 +200MB） |

推荐使用根目录的 Docker Compose 一键部署：

```bash
# PostgreSQL 模式（带独立数据库容器）
docker compose up -d

# SQLite 模式（无需数据库容器，更轻量）
docker compose -f docker-compose.sqlite.yml up -d

# 启用网页快照功能（安装 Chromium）
INCLUDE_CHROME=1 docker compose up -d
```

## 💻 开发说明

### 添加新的 API 端点

1. 在 `handlers/` 目录创建或修改处理器文件
2. 在 `routes/routes.go` 中注册路由，按需挂载认证 / 管理员中间件
3. 如需新数据模型，在 `models/models.go` 中添加并在 `internal/database.go` 注册迁移

### 生产环境注意事项

1. 修改 `JWT_SECRET` 为强随机字符串（默认值会被拒绝启动）
2. 配置正确的 `DATABASE_URL` 或切换 `DB_DRIVER=sqlite`
3. 设置正确的 `FRONTEND_URL`（CORS）
4. 启用 HTTPS
5. 配置日志与监控

## 📝 许可证

MIT License
