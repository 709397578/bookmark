# Bookmark API 测试指南

## 📋 目录
- [快速开始](#快速开始)
- [认证 API](#认证-api)
- [收藏夹 API](#收藏夹-api)
- [书签 API](#书签-api)
- [文件夹 API](#文件夹-api)
- [使用 Apifox/Postman](#使用-apifoxpostman)

---

## 🚀 快速开始

### 为什么返回空数据？

当你访问 `http://localhost:8080/api/bookmarks` 返回空数据时，这是**正常的**，因为：
1. ✅ API 工作正常
2. ✅ 数据库连接成功
3. ❌ 但数据库中还没有任何数据

**解决方案：** 需要先创建用户、收藏夹，然后才能添加书签。

---

## 🔐 认证 API

### 1. 注册用户

**请求：**
```http
POST http://localhost:8080/api/auth/register
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "123456",
  "name": "Test User"
}
```

**响应：**
```json
{
  "success": true,
  "message": "注册成功",
  "data": {
    "id": "uuid-here",
    "email": "test@example.com",
    "name": "Test User"
  }
}
```

### 2. 登录获取 Token

**请求：**
```http
POST http://localhost:8080/api/auth/login
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "123456"
}
```

**响应：**
```json
{
  "success": true,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "uuid-here",
      "email": "test@example.com",
      "name": "Test User",
      "role": "user"
    }
  }
}
```

⚠️ **重要：** 复制 `token` 值，后续请求需要使用它！

---

## 📁 收藏夹 API

### 3. 创建收藏夹

**请求：**
```http
POST http://localhost:8080/api/collections
Content-Type: application/json
Authorization: Bearer YOUR_TOKEN_HERE

{
  "name": "我的书签",
  "description": "个人收藏的网址",
  "icon": "🔖",
  "isPublic": true,
  "sortOrder": 0
}
```

**响应：**
```json
{
  "success": true,
  "message": "创建成功",
  "data": {
    "id": "collection-uuid-here",
    "name": "我的书签",
    "slug": "我的书签-1234567890",
    "description": "个人收藏的网址",
    "icon": "🔖",
    "isPublic": true,
    "sortOrder": 0,
    "userId": "user-uuid",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

⚠️ **重要：** 复制 `id` 值（collection-uuid-here），创建书签时需要用到！

### 4. 获取收藏夹列表

**请求：**
```http
GET http://localhost:8080/api/collections?publicOnly=true
```

**响应：**
```json
{
  "success": true,
  "data": [
    {
      "id": "collection-uuid",
      "name": "我的书签",
      "slug": "我的书签-1234567890",
      "description": "个人收藏的网址",
      "icon": "🔖",
      "isPublic": true,
      "sortOrder": 0
    }
  ]
}
```

---

## 🔖 书签 API

### 5. 创建书签

**请求：**
```http
POST http://localhost:8080/api/bookmarks
Content-Type: application/json
Authorization: Bearer YOUR_TOKEN_HERE

{
  "title": "GitHub",
  "url": "https://github.com",
  "description": "代码托管平台",
  "icon": "https://github.com/favicon.ico",
  "collectionId": "YOUR_COLLECTION_ID_HERE",
  "isFeatured": false,
  "sortOrder": 0
}
```

**参数说明：**
- `title`: 书签标题（必填）
- `url`: 书签URL（必填）
- `description`: 描述（可选）
- `icon`: 图标URL（可选）
- `collectionId`: 收藏夹ID（必填）
- `folderId`: 文件夹ID（可选）
- `isFeatured`: 是否精选（可选，默认false）
- `sortOrder`: 排序（可选，默认0）

**响应：**
```json
{
  "success": true,
  "message": "创建成功",
  "data": {
    "id": "bookmark-uuid",
    "title": "GitHub",
    "url": "https://github.com",
    "description": "代码托管平台",
    "icon": "https://github.com/favicon.ico",
    "collectionId": "collection-uuid",
    "isFeatured": false,
    "sortOrder": 0,
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
}
```

### 6. 获取书签列表

**请求：**
```http
GET http://localhost:8080/api/bookmarks?page=1&pageSize=100
```

**带筛选条件：**
```http
GET http://localhost:8080/api/bookmarks?collectionId=YOUR_COLLECTION_ID&page=1&pageSize=10
```

**响应：**
```json
{
  "success": true,
  "data": {
    "data": [
      {
        "id": "bookmark-uuid",
        "title": "GitHub",
        "url": "https://github.com",
        "description": "代码托管平台",
        "icon": "https://github.com/favicon.ico",
        "collectionId": "collection-uuid",
        "isFeatured": false,
        "sortOrder": 0
      }
    ],
    "total": 1,
    "currentPage": 1,
    "totalPages": 1,
    "pageSize": 100
  }
}
```

---

## 📂 文件夹 API（可选）

### 7. 创建文件夹

**请求：**
```http
POST http://localhost:8080/api/folders
Content-Type: application/json
Authorization: Bearer YOUR_TOKEN_HERE

{
  "name": "开发工具",
  "collectionId": "YOUR_COLLECTION_ID_HERE",
  "sortOrder": 0
}
```

### 8. 创建带文件夹的书签

**请求：**
```http
POST http://localhost:8080/api/bookmarks
Content-Type: application/json
Authorization: Bearer YOUR_TOKEN_HERE

{
  "title": "VS Code",
  "url": "https://code.visualstudio.com",
  "collectionId": "YOUR_COLLECTION_ID_HERE",
  "folderId": "YOUR_FOLDER_ID_HERE"
}
```

---

## 🛠️ 使用 Apifox/Postman

### 配置步骤

#### 1. 创建环境变量

在 Apifox/Postman 中创建以下环境变量：

| 变量名 | 值 | 说明 |
|--------|-----|------|
| `base_url` | `http://localhost:8080` | API 基础地址 |
| `token` | _(登录后填写)_ | JWT Token |
| `collection_id` | _(创建后填写)_ | 收藏夹ID |
| `folder_id` | _(创建后填写)_ | 文件夹ID |

#### 2. 设置请求头

在所有需要认证的请求中添加：

```
Authorization: Bearer {{token}}
Content-Type: application/json
```

#### 3. 测试流程

**步骤 1：注册用户**
```
POST {{base_url}}/api/auth/register
Body: {"email":"test@example.com","password":"123456","name":"Test"}
```

**步骤 2：登录获取 Token**
```
POST {{base_url}}/api/auth/login
Body: {"email":"test@example.com","password":"123456"}
```
→ 复制响应中的 `token` 值到环境变量

**步骤 3：创建收藏夹**
```
POST {{base_url}}/api/collections
Headers: Authorization: Bearer {{token}}
Body: {"name":"测试收藏夹","isPublic":true}
```
→ 复制响应中的 `id` 到 `collection_id` 环境变量

**步骤 4：创建书签**
```
POST {{base_url}}/api/bookmarks
Headers: Authorization: Bearer {{token}}
Body: {
  "title":"GitHub",
  "url":"https://github.com",
  "collectionId":"{{collection_id}}"
}
```

**步骤 5：获取书签列表**
```
GET {{base_url}}/api/bookmarks?collectionId={{collection_id}}
```
→ 现在应该能看到数据了！

---

## 📝 完整测试示例

### 一次性添加多个测试书签

你可以批量创建书签来测试：

```json
// 书签 1: GitHub
{
  "title": "GitHub",
  "url": "https://github.com",
  "description": "全球最大的代码托管平台",
  "collectionId": "YOUR_COLLECTION_ID"
}

// 书签 2: Stack Overflow
{
  "title": "Stack Overflow",
  "url": "https://stackoverflow.com",
  "description": "程序员问答社区",
  "collectionId": "YOUR_COLLECTION_ID"
}

// 书签 3: MDN Web Docs
{
  "title": "MDN Web Docs",
  "url": "https://developer.mozilla.org",
  "description": "Web 技术文档",
  "collectionId": "YOUR_COLLECTION_ID"
}

// 书签 4: Next.js
{
  "title": "Next.js",
  "url": "https://nextjs.org",
  "description": "React 框架",
  "collectionId": "YOUR_COLLECTION_ID"
}

// 书签 5: Go 语言
{
  "title": "Go Programming Language",
  "url": "https://go.dev",
  "description": "Go 语言官方网站",
  "collectionId": "YOUR_COLLECTION_ID"
}
```

---

## 🔍 常见问题

### Q1: 为什么返回 401 Unauthorized？
**A:** Token 无效或已过期。请重新登录获取新 Token。

### Q2: 为什么返回 400 Bad Request？
**A:** 请求参数错误。检查必填字段是否完整：
- 创建书签需要：`title`, `url`, `collectionId`
- 创建收藏夹需要：`name`

### Q3: 如何查看数据库中有哪些数据？
**A:** 可以直接查询数据库：
```sql
-- 查看所有用户
SELECT * FROM users;

-- 查看所有收藏夹
SELECT * FROM collections;

-- 查看所有书签
SELECT * FROM bookmarks;

-- 统计数量
SELECT COUNT(*) FROM bookmarks;
```

### Q4: 如何删除测试数据？
**A:** 
```sql
-- 删除所有书签
DELETE FROM bookmarks;

-- 删除所有收藏夹
DELETE FROM collections;

-- 删除所有用户
DELETE FROM users;
```

---

PowerShell 自动化测试脚本，一键添加测试数据
```powershell
cd backend
powershell -ExecutionPolicy Bypass -File test-api.ps1
```

## 🎯 总结

**正确的测试流程：**
1. ✅ 注册用户
2. ✅ 登录获取 Token
3. ✅ 创建收藏夹
4. ✅ 创建书签（使用收藏夹ID）
5. ✅ 查询书签列表

现在你应该能看到数据了！如果还有问题，请检查：
- Token 是否正确
- Collection ID 是否正确
- 请求头是否包含 `Authorization: Bearer <token>`
