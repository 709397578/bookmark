import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { authAPI } from '@/api'

interface User {
  id: string
  email: string
  name: string
  role: string
  avatar?: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token'))

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  // 初始化用户信息
  function initAuth() {
    const savedUser = localStorage.getItem('user')
    const savedToken = localStorage.getItem('token')
    if (savedUser && savedToken) {
      user.value = JSON.parse(savedUser)
      token.value = savedToken
    }
  }

  // 登录
  async function login(email: string, password: string) {
    console.log('[authStore.login] 开始登录', { email })
    try {
      console.log('[authStore.login] 调用 API...')
      const response: any = await authAPI.login({ email, password })
      console.log('[authStore.login] API 响应:', response)

      // 后端返回格式: { success: true, message: "...", data: {...} }
      if (response.success === true) {
        console.log('[authStore.login] 登录成功，保存数据')
        // 先更新localStorage
        localStorage.setItem('token', response.data.token)
        localStorage.setItem('user', JSON.stringify(response.data.user))

        // 再更新store状态
        token.value = response.data.token
        user.value = response.data.user

        console.log('[authStore.login] 状态已更新')
        return { success: true, message: response.message }
      }
      console.log('[authStore.login] 登录失败:', response.message || response.error)
      return { success: false, message: response.message || response.error || '登录失败' }
    } catch (error: any) {
      console.error('[authStore.login] 捕获异常:', error)
      return {
        success: false,
        message:
          error.response?.data?.message ||
          error.response?.data?.error ||
          '登录失败，请检查网络连接',
      }
    }
  }

  // 注册
  async function register(email: string, password: string, name?: string) {
    console.log('[authStore.register] 开始注册', { email })
    try {
      console.log('[authStore.register] 调用 API...')
      const response: any = await authAPI.register({ email, password, name })
      console.log('[authStore.register] API 响应:', response)

      // 后端返回格式: { success: true, message: "...", data: {...} }
      if (response.success === true) {
        console.log('[authStore.register] 注册成功')
        return { success: true, message: response.message }
      }
      console.log('[authStore.register] 注册失败:', response.message || response.error)
      return { success: false, message: response.message || response.error || '注册失败' }
    } catch (error: any) {
      console.error('[authStore.register] 捕获异常:', error)
      return {
        success: false,
        message:
          error.response?.data?.message ||
          error.response?.data?.error ||
          '注册失败，请检查网络连接',
      }
    }
  }

  // 登出
  function logout() {
    user.value = null
    token.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return {
    user,
    token,
    isAuthenticated,
    isAdmin,
    initAuth,
    login,
    register,
    logout,
  }
})
