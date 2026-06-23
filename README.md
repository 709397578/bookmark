# Bookmart 前端

PinTree 是一个现代化的书签管理系统，支持书签收藏、文件夹分类、网页快照、导入导出等功能。本仓库根目录为前端项目（Vue 3 + Vite），后端代码位于 [`backend/`](./backend)。

> 详细的架构设计参见 [ARCHITECTURE.md](./ARCHITECTURE.md)
## 项目预览
<img width="1440" height="621" alt="image" src="https://github.com/user-attachments/assets/3cd76394-00c3-4d4d-aa2f-e6ad6caea8d7" />
<img width="1250" height="498" alt="image" src="https://github.com/user-attachments/assets/5bba23ea-767f-4bd7-877b-0e282699f3b0" />


## ✨ 功能特性

- 📚 **收藏夹管理**：创建、编辑、删除收藏夹，支持拖拽排序
- 📁 **文件夹分类**：多级文件夹组织书签，支持颜色与图标
- 🔖 **书签管理**：增删改查、分页浏览、关键词搜索、批量移动/删除/排序
- 📸 **网页快照**：调用后端生成并预览网页快照
- 📤 **导入导出**：支持浏览器书签 HTML 格式的导入与导出
- 👤 **用户系统**：登录/注册、头像上传、修改密码
- 🛡️ **后台管理**：用户管理、站点设置（管理员）
- 📱 **响应式布局**：桌面端 / 移动端自适应

## 🛠️ 技术栈

| 类别 | 技术 |
| --- | --- |
| 框架 | Vue 3（`<script setup>` + TypeScript） |
| 构建工具 | Vite |
| 路由 | Vue Router |
| 状态管理 | Pinia |
| UI 组件库 | Element Plus + Vant（移动端） |
| HTTP 客户端 | Axios |
| 拖拽 | SortableJS |
| 代码规范 | Prettier |

## 📁 项目结构

```
src/
├── api/           # Axios 封装与接口定义
│   └── index.ts   # authAPI / collectionAPI / folderAPI / bookmarkAPI / settingAPI
├── assets/        # 静态资源（样式、图片）
├── components/    # 公共组件
│   ├── AppLayout.vue        # 整体布局
│   └── DesktopLayout.vue    # 桌面端布局
├── router/        # 路由配置与全局守卫
│   └── index.ts
├── stores/        # Pinia 状态管理
│   ├── auth.ts        # 认证状态
│   ├── collection.ts  # 收藏夹/书签业务状态
│   └── counter.ts
├── utils/         # 工具函数（device.ts / ui.ts）
├── views/         # 页面视图
│   ├── LoginView.vue          # 登录/注册
│   ├── HomeView.vue           # 首页（书签浏览，公开可访问）
│   ├── AboutView.vue          # 关于页
│   ├── BookmarkManageView.vue # 书签管理
│   ├── BookmarkEditView.vue   # 书签新增/编辑
│   ├── BookmarkImportView.vue # 书签导入
│   ├── FolderManageView.vue   # 文件夹管理
│   ├── UserManageView.vue     # 用户管理（管理员）
│   ├── SettingsView.vue       # 站点设置（管理员）
│   └── ...
├── App.vue        # 根组件
└── main.ts        # 应用入口
```

## 🚀 快速开始

### 环境要求

- Node.js `^20.19.0 || >=22.12.0`
- 后端服务已启动（默认 `http://localhost:8080`，参见 [`backend/README.md`](./backend/README.md)）

### 安装与运行

```sh
# 安装依赖
npm install

# 启动开发服务器（默认 http://localhost:5173）
npm run dev
```

开发服务器已配置代理，将 `/api`、`/snapshots`、`/avatars` 请求转发至后端（见 `vite.config.ts`），因此前端可直接使用 `/api` 作为请求基址，无需关心跨域。

### 构建与预览

```sh
# 类型检查 + 生产构建（产物输出到 dist/）
npm run build

# 本地预览生产构建
npm run preview
```

### 代码格式化

```sh
npm run format
```

## 🔐 认证机制

- 登录成功后将 `token` 与 `user` 存入 `localStorage`
- Axios 请求拦截器自动在请求头附加 `Authorization: Bearer <token>`
- 响应拦截器在收到 `401`（非登录/注册接口）时清除凭证并跳转登录页
- 路由守卫根据 `meta.requiresAuth` 判断页面是否需要登录

## 🌐 路由概览

| 路径 | 页面 | 是否需要登录 |
| --- | --- | --- |
| `/login` | 登录/注册 | 否 |
| `/` | 首页（书签浏览） | 否（公开可访问） |
| `/about` | 关于 | 否 |
| `/bookmarks` | 书签管理 | 是 |
| `/bookmark/add` | 新增书签 | 是 |
| `/bookmark/edit/:id` | 编辑书签 | 是 |
| `/bookmarks/import` | 书签导入 | 是 |
| `/folders` | 文件夹管理 | 是 |
| `/admin/users` | 用户管理 | 是（管理员） |
| `/admin/settings` | 站点设置 | 是（管理员） |

## 🐳 Docker 部署

前端 Dockerfile（`Dockerfile.frontend`）使用多阶段构建：Node 编译 + Nginx 托管静态文件。

```sh
# 通过根目录 docker-compose 一键启动（含后端 + 数据库）
docker compose up -d            # PostgreSQL 模式
# 或使用 SQLite 模式（无需数据库容器）
docker compose -f docker-compose.sqlite.yml up -d
```

生产环境通过 `nginx.conf` 反向代理 `/api` 至后端容器，并托管 `dist/` 静态资源。

## 快捷部署：使用docker-compose.yml一键部署
```yml
services:
  backend:
    image: ccr.ccs.tencentyun.com/bookmark/bookmark-backend:0.1.8
    container_name: bookmark-backend
    environment:
      PORT: 8080
      DB_DRIVER: sqlite
      DB_PATH: /data/pintree.db
      JWT_SECRET: ${JWT_SECRET:-your-production-jwt-secret}
      JWT_EXPIRATION: 86400
      FRONTEND_URL: http://localhost
      UPLOAD_PATH: /root/uploads
    ports:
      - '8085:8080'
    volumes:
      - ./sqlite_data:/data
      - ./upload_data:/root/uploads
    restart: unless-stopped

  frontend:
    image: ccr.ccs.tencentyun.com/bookmark/bookmark-frontend:0.1.8
    container_name: bookmark-frontend
    ports:
      - '80:80'
    depends_on:
      - backend
    restart: unless-stopped
```

## 📦 环境变量

前端本身无需 `.env` 配置，开发期通过 Vite 代理与后端通信。Nginx 部署时相关反代规则见 `nginx.conf`。

## 📜 可用脚本

| 命令 | 说明 |
| --- | --- |
| `npm run dev` | 启动开发服务器（热更新） |
| `npm run build` | 类型检查 + 生产构建 |
| `npm run preview` | 预览生产构建 |
| `npm run type-check` | 仅执行 `vue-tsc` 类型检查 |
| `npm run format` | Prettier 格式化 `src/` |

## 📝 许可证

MIT License
