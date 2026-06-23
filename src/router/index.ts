import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      name: 'home',
      component: () => import('../views/HomeView.vue'),
      meta: { requiresAuth: false }, // 允许未登录访问
    },
    {
      path: '/about',
      name: 'about',
      component: () => import('../views/AboutView.vue'),
      meta: { requiresAuth: false }, // 允许未登录访问
    },
    {
      path: '/bookmarks',
      name: 'bookmarks',
      component: () => import('../views/BookmarkManageView.vue'),
      meta: { requiresAuth: true }, // 需要登录才能访问
    },
    {
      path: '/bookmark/add',
      name: 'bookmark-add',
      component: () => import('../views/BookmarkEditView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/bookmark/edit/:id',
      name: 'bookmark-edit',
      component: () => import('../views/BookmarkEditView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/folders',
      name: 'folder-manage',
      component: () => import('../views/FolderManageView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/bookmarks/import',
      name: 'bookmark-import',
      component: () => import('../views/BookmarkImportView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/admin/users',
      name: 'user-manage',
      component: () => import('../views/UserManageView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/admin/settings',
      name: 'settings',
      component: () => import('../views/SettingsView.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

// 路由守卫
router.beforeEach((to, from) => {
  const authStore = useAuthStore()

  // 确保在每次路由切换前初始化认证状态
  if (!authStore.token && !authStore.user) {
    authStore.initAuth()
  }

  // 如果已登录且访问登录页，跳转到首页
  if (to.path === '/login' && authStore.isAuthenticated) {
    return '/'
  }

  // 如果页面需要认证且未登录，跳转到登录页
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }

  // 继续当前导航
  return true
})

export default router
