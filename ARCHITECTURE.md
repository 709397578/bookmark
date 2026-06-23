# Bookmark 前端架构设计文档

## 🏗️ 整体架构

```
┌─────────────────────────────────────────────────┐
│                  用户界面层 (UI)                   │
├─────────────────────────────────────────────────┤
│  LoginView  │  HomeView  │  AboutView           │
└─────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────┐
│                组件层 (Components)                 │
├─────────────────────────────────────────────────┤
│              AppLayout (主布局)                    │
│  ┌──────────────┐    ┌──────────────────────┐   │
│  │   Sidebar    │    │    Main Content      │   │
│  │  - 用户信息   │    │  - Top Bar (搜索)     │   │
│  │  - 收藏夹列表 │    │  - Content Area      │   │
│  │  - 文件夹列表 │    │    (RouterView)      │   │
│  └──────────────┘    └──────────────────────┘   │
└─────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────┐
│              状态管理层 (Pinia Stores)             │
├─────────────────────────────────────────────────┤
│  ┌──────────────┐         ┌──────────────────┐  │
│  │  authStore   │         │ collectionStore  │  │
│  ├──────────────┤         ├──────────────────┤  │
│  │ • user       │         │ • collections    │  │
│  │ • token      │         │ • currentColl.   │  │
│  │ • login()    │         │ • folders        │  │
│  │ • register() │         │ • bookmarks      │  │
│  │ • logout()   │         │ • CRUD methods   │  │
│  └──────────────┘         └──────────────────┘  │
└─────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────┐
│               API层 (Axios封装)                    │
├─────────────────────────────────────────────────┤
│  ┌────────────┐ ┌──────────┐ ┌──────────────┐  │
│  │ authAPI    │ │collAPI   │ │bookmarkAPI   │  │
│  ├────────────┤ ├──────────┤ ├──────────────┤  │
│  │ • login    │ │• getList │ │• getList     │  │
│  │ • register │ │• create  │ │• create      │  │
│  └────────────┘ │• update  │ │• update      │  │
│                 │• delete  │ │• delete      │  │
│                 └──────────┘ │• search      │  │
│                              └──────────────┘  │
│                                                │
│  拦截器:                                        │
│  • Request: 自动添加 Token                      │
│  • Response: 401自动跳转登录                     │
└─────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────┐
│            路由层 (Vue Router)                     │
├─────────────────────────────────────────────────┤
│  /login  →  LoginView  (公开)                    │
│  /       →  HomeView   (需认证)                   │
│  /about  →  AboutView  (需认证)                   │
│                                                │
│  路由守卫: 检查认证状态，自动重定向                  │
└─────────────────────────────────────────────────┘
```

## 📦 数据流设计

### 1. 用户登录流程
```
用户输入 → LoginView
    ↓
调用 authStore.login()
    ↓
发送 POST /api/auth/login
    ↓
后端验证 → 返回 Token + User
    ↓
存储到 localStorage
    ↓
更新 authStore 状态
    ↓
路由跳转到首页 (/)
```

### 2. 获取收藏夹流程
```
页面加载 → HomeView.onMounted()
    ↓
调用 collectionStore.fetchCollections()
    ↓
发送 GET /api/collections
    ↓
后端返回收藏夹列表
    ↓
更新 collectionStore.collections
    ↓
渲染侧边栏收藏夹列表
```

### 3. 添加书签流程
```
用户点击"添加书签" → 打开对话框
    ↓
填写表单 → 提交
    ↓
调用 collectionStore.createBookmark()
    ↓
发送 POST /api/bookmarks
    ↓
后端创建书签 → 返回新书签
    ↓
添加到 collectionStore.bookmarks
    ↓
关闭对话框 → 显示成功提示
    ↓
UI自动更新（响应式）
```

### 4. 搜索书签流程
```
用户输入关键词 → 点击搜索
    ↓
调用 collectionStore.searchBookmarks(query)
    ↓
发送 GET /api/bookmarks/search?q=xxx
    ↓
后端返回匹配的书签
    ↓
更新搜索结果
    ↓
显示结果数量提示
```

## 🎯 核心设计模式

### 1. 组合式API (Composition API)
```typescript
// 每个组件使用 <script setup>
<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

// 状态
const authStore = useAuthStore()

// 方法
const handleLogin = async () => {
  // ...
}

// 生命周期
onMounted(() => {
  // ...
})
</script>
```

### 2. Store模块化设计
```typescript
// 按功能划分Store
authStore       → 用户认证相关
collectionStore → 业务数据相关

// 每个Store包含：
- State (响应式状态)
- Getters (计算属性)
- Actions (异步操作)
```

### 3. API分层设计
```typescript
// 第一层：Axios实例配置
const api = axios.create({ baseURL, timeout })

// 第二层：接口分组
export const authAPI = { login, register }
export const collectionAPI = { ... }
export const bookmarkAPI = { ... }

// 第三层：Store调用API
async function fetchCollections() {
  const response = await collectionAPI.getCollections()
  // 处理响应，更新状态
}
```

### 4. 组件通信模式
```
Parent Component
    ↓ Props
Child Component
    ↓ Emits
Parent Component

通过Store共享状态（跨组件通信）
```

## 🔐 安全设计

### 1. Token管理
```typescript
// 存储
localStorage.setItem('token', token)

// 读取（请求拦截器）
const token = localStorage.getItem('token')
config.headers.Authorization = `Bearer ${token}`

// 清除（登出或401）
localStorage.removeItem('token')
```

### 2. 路由保护
```typescript
// 路由元信息
meta: { requiresAuth: true }

// 全局守卫
router.beforeEach((to, from, next) => {
  if (requiresAuth && !isAuthenticated) {
    next('/login')
  } else {
    next()
  }
})
```

### 3. XSS防护
- Vue自动转义输出
- 避免使用 v-html
- URL验证和清理

## 🎨 UI/UX设计原则

### 1. 视觉层次
```
一级标题 (h1)  → 32px, 粗体
二级标题 (h2)  → 24px, 中粗体
正文文本       → 14-16px, 常规
辅助文本       → 12px, 浅色
```

### 2. 色彩系统
```
主色: #1890ff (蓝色)
成功: #52c41a (绿色)
警告: #faad14 (橙色)
错误: #ff4d4f (红色)
文字: #333/#666/#999 (深/中/浅灰)
背景: #f5f5f5 (浅灰)
```

### 3. 间距系统
```
xs:  4px
sm:  8px
md:  16px
lg:  24px
xl:  32px
xxl: 48px
```

### 4. 圆角规范
```
小圆角: 4px  (按钮、标签)
中圆角: 8px  (卡片、输入框)
大圆角: 12px (对话框、容器)
超大:   16px (登录框)
```

## 📱 响应式设计

### 断点设置
```css
移动端: < 768px
平板:   768px - 1024px
桌面:   > 1024px
```

### 布局适配
```
桌面端:
- 侧边栏: 280px (可折叠到60px)
- 书签网格: 3-4列

平板端:
- 侧边栏: 240px
- 书签网格: 2-3列

移动端:
- 侧边栏: 抽屉式
- 书签网格: 1-2列
```

## ⚡ 性能优化

### 1. 代码分割
```typescript
// 路由懒加载
component: () => import('../views/LoginView.vue')
```

### 2. 组件缓存
```typescript
// KeepAlive缓存（可选）
<KeepAlive>
  <RouterView />
</KeepAlive>
```

### 3. 防抖节流
```typescript
// 搜索防抖（建议实现）
import { debounce } from 'lodash-es'
const handleSearch = debounce((query) => {
  // 搜索逻辑
}, 300)
```

### 4. 图片优化
- 使用emoji代替图标图片
- 懒加载（大数据量时）
- WebP格式支持

## 🧪 测试策略

### 单元测试（建议）
```typescript
// Store测试
describe('authStore', () => {
  it('should login successfully', () => {
    // 测试登录逻辑
  })
})

// API测试
describe('collectionAPI', () => {
  it('should fetch collections', async () => {
    // 测试API调用
  })
})
```

### E2E测试（建议）
```typescript
// Cypress测试
describe('Login Flow', () => {
  it('should login and redirect to home', () => {
    cy.visit('/login')
    cy.get('[name="email"]').type('test@example.com')
    cy.get('[name="password"]').type('password')
    cy.get('button[type="submit"]').click()
    cy.url().should('include', '/')
  })
})
```

## 🔄 状态同步机制

### 乐观更新
```typescript
// 删除操作
async function deleteBookmark(id: string) {
  // 1. 立即从UI移除
  bookmarks.value = bookmarks.value.filter(b => b.id !== id)
  
  try {
    // 2. 发送请求
    await bookmarkAPI.deleteBookmark(id)
    showToast('删除成功')
  } catch (error) {
    // 3. 失败则恢复
    fetchBookmarks() // 重新获取
    showToast('删除失败')
  }
}
```

### 悲观更新
```typescript
// 创建操作
async function createBookmark(data) {
  loading.value = true
  try {
    // 1. 等待服务器响应
    const response = await bookmarkAPI.createBookmark(data)
    
    // 2. 成功后更新UI
    bookmarks.value.unshift(response.data)
    showToast('创建成功')
  } catch (error) {
    showToast('创建失败')
  } finally {
    loading.value = false
  }
}
```

## 📊 错误处理

### 统一错误处理
```typescript
// API拦截器
api.interceptors.response.use(
  response => response.data,
  error => {
    if (error.response?.status === 401) {
      // 未授权，跳转登录
      authStore.logout()
      router.push('/login')
    }
    return Promise.reject(error)
  }
)

// 组件级错误处理
try {
  await someAction()
} catch (error) {
  showToast(error.message || '操作失败')
}
```

## 🎓 最佳实践

### 1. 命名规范
```typescript
// Store: useXxxStore
useAuthStore, useCollectionStore

// API: xxxAPI
authAPI, collectionAPI

// 组件: XxxView / XxxComponent
LoginView, AppLayout

// 变量: camelCase
const userName = ref('')

// 常量: UPPER_SNAKE_CASE
const MAX_PAGE_SIZE = 100
```

### 2. 文件组织
```
src/
├── api/          # API接口
├── assets/       # 静态资源
├── components/   # 公共组件
├── router/       # 路由配置
├── stores/       # 状态管理
├── views/        # 页面视图
├── App.vue       # 根组件
└── main.ts       # 入口文件
```

### 3. 注释规范
```typescript
/**
 * 获取收藏夹列表
 * @param publicOnly - 是否只获取公开的收藏夹
 * @returns 收藏夹数组
 */
async function fetchCollections(publicOnly?: boolean) {
  // 实现...
}
```

---

**架构设计理念**: 简洁、清晰、可维护、可扩展 🚀
