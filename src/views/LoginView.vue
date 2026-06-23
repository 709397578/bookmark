<template>
  <div class="login-container">
    <div class="login-box">
      <h1 class="login-title">📚 PinTree</h1>
      <p class="login-subtitle">书签管理系统</p>

      <!-- 登录表单 -->
      <van-form v-if="isLogin" @submit="handleLogin" :class="{ shake: shaking }">
        <van-cell-group inset>
          <van-field
            v-model="loginForm.email"
            name="email"
            label="邮箱"
            placeholder="请输入邮箱"
            :rules="[{ required: true, message: '请填写邮箱' }]"
          />
          <van-field
            v-model="loginForm.password"
            type="password"
            name="password"
            label="密码"
            placeholder="请输入密码"
            :rules="[{ required: true, message: '请填写密码' }]"
          />
        </van-cell-group>

        <transition name="slide-fade">
          <div v-if="formError" class="error-banner">
            <span class="error-icon">!</span>
            <span>{{ formError }}</span>
          </div>
        </transition>

        <div class="button-group">
          <van-button round block type="primary" native-type="submit" :loading="loading">
            登录
          </van-button>
          <van-button v-if="allowRegistration" round block plain type="primary" @click="toggleMode">
            去注册
          </van-button>
        </div>
      </van-form>

      <!-- 注册表单 -->
      <van-form v-else-if="allowRegistration" @submit="handleRegister" :class="{ shake: shaking }">
        <van-cell-group inset>
          <van-field
            v-model="registerForm.name"
            name="name"
            label="姓名"
            placeholder="请输入姓名（可选）"
          />
          <van-field
            v-model="registerForm.email"
            name="email"
            label="邮箱"
            placeholder="请输入邮箱"
            :rules="[
              { required: true, message: '请填写邮箱' },
              { pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/, message: '邮箱格式不正确' },
            ]"
          />
          <van-field
            v-model="registerForm.password"
            type="password"
            name="password"
            label="密码"
            placeholder="请输入密码（至少8位，需包含字母和数字）"
            :rules="[
              { required: true, message: '请填写密码' },
              { validator: validatePassword, message: '密码至少8位，需包含字母和数字' },
            ]"
          />
          <van-field
            v-model="registerForm.confirmPassword"
            type="password"
            name="confirmPassword"
            label="确认密码"
            placeholder="请再次输入密码"
            :rules="[
              { required: true, message: '请确认密码' },
              { validator: validateConfirmPassword, message: '两次密码不一致' },
            ]"
          />
        </van-cell-group>

        <transition name="slide-fade">
          <div v-if="formError" class="error-banner">
            <span class="error-icon">!</span>
            <span>{{ formError }}</span>
          </div>
        </transition>

        <div class="button-group">
          <van-button round block type="primary" native-type="submit" :loading="loading">
            注册
          </van-button>
          <van-button round block plain type="primary" @click="toggleMode"> 去登录 </van-button>
        </div>
      </van-form>

      <!-- 注册已关闭时显示 -->
      <div v-else-if="!isLogin && !allowRegistration" class="registration-closed">
        <div class="button-group">
          <van-button round block plain type="primary" @click="toggleMode"> 去登录 </van-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { useAuthStore } from '@/stores/auth'
import { settingAPI } from '@/api'

const router = useRouter()
const authStore = useAuthStore()

const isLogin = ref(true)
const loading = ref(false)
const formError = ref('')
const shaking = ref(false)
const allowRegistration = ref(true)

onMounted(async () => {
  try {
    const response: any = await settingAPI.getPublicSettings()
    if (response?.code === 200 || response?.success) {
      if (typeof response.data?.allowRegistration === 'boolean') {
        allowRegistration.value = response.data.allowRegistration
      }
    }
  } catch {
    // 获取失败默认允许注册
  }
})

let errorTimer: ReturnType<typeof setTimeout> | null = null

const showError = (msg: string) => {
  if (errorTimer) clearTimeout(errorTimer)
  formError.value = msg
  shaking.value = true
  setTimeout(() => {
    shaking.value = false
  }, 500)
  errorTimer = setTimeout(() => {
    formError.value = ''
    errorTimer = null
  }, 4000)
}

const clearError = () => {
  if (errorTimer) clearTimeout(errorTimer)
  formError.value = ''
  shaking.value = false
  errorTimer = null
}

const loginForm = reactive({
  email: '',
  password: '',
})

const registerForm = reactive({
  name: '',
  email: '',
  password: '',
  confirmPassword: '',
})

// 切换登录/注册模式
const toggleMode = async () => {
  clearError()
  // 切换到注册模式时重新检查是否允许注册
  if (isLogin.value) {
    try {
      const response: any = await settingAPI.getPublicSettings()
      if (response?.success && typeof response.data?.allowRegistration === 'boolean') {
        allowRegistration.value = response.data.allowRegistration
      }
    } catch {
      // 获取失败沿用当前值
    }
  }
  isLogin.value = !isLogin.value
  // 清空表单
  loginForm.email = ''
  loginForm.password = ''
  registerForm.name = ''
  registerForm.email = ''
  registerForm.password = ''
  registerForm.confirmPassword = ''
}

// 验证密码强度（必须与后端规则一致）
const validatePassword = (val: string) => {
  if (val.length < 8) return false
  if (!/[a-zA-Z]/.test(val)) return false
  if (!/[0-9]/.test(val)) return false
  return true
}

// 验证确认密码
const validateConfirmPassword = (val: string) => {
  return val === registerForm.password
}

// 处理登录
const handleLogin = async () => {
  clearError()
  loading.value = true

  try {
    const result = await authStore.login(loginForm.email, loginForm.password)

    if (result.success) {
      showToast({ message: '登录成功', icon: 'success' })
      await new Promise((resolve) => setTimeout(resolve, 100))
      router.replace('/')
    } else {
      showError(result.message || '登录失败')
    }
  } catch (error) {
    showError('登录失败，请检查网络连接')
  } finally {
    loading.value = false
  }
}

// 处理注册
const handleRegister = async () => {
  clearError()
  loading.value = true
  try {
    const result = await authStore.register(
      registerForm.email,
      registerForm.password,
      registerForm.name,
    )
    if (result.success) {
      showToast({ message: '注册成功，请登录', icon: 'success' })
      toggleMode()
    } else {
      showError(result.message || '注册失败')
    }
  } catch (error) {
    showError('注册失败，请检查网络连接')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-box {
  width: 100%;
  max-width: 400px;
  background: white;
  border-radius: 16px;
  padding: 40px 30px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
}

.login-title {
  text-align: center;
  font-size: 32px;
  margin-bottom: 8px;
  color: #333;
}

.login-subtitle {
  text-align: center;
  font-size: 14px;
  color: #999;
  margin-bottom: 30px;
}

.button-group {
  margin-top: 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 错误横幅 */
.error-banner {
  margin-top: 16px;
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  color: #b91c1c;
  border: 1px solid #fecaca;
  border-radius: 10px;
  padding: 10px 14px;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 10px;
  backdrop-filter: blur(4px);
}

.error-icon {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #ef4444;
  color: #fff;
  font-size: 12px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

/* 注册已关闭提示 */
.registration-closed {
  min-height: 100px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 抖动动画 */
.shake {
  animation: shake 0.4s ease-in-out;
}

@keyframes shake {
  0%,
  100% {
    transform: translateX(0);
  }
  20% {
    transform: translateX(-8px);
  }
  40% {
    transform: translateX(8px);
  }
  60% {
    transform: translateX(-5px);
  }
  80% {
    transform: translateX(4px);
  }
}

/* 错误横幅进出动画 */
.slide-fade-enter-active {
  transition: all 0.3s ease-out;
}

.slide-fade-leave-active {
  transition: all 0.25s ease-in;
}

.slide-fade-enter-from {
  opacity: 0;
  transform: translateY(-6px);
}

.slide-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
