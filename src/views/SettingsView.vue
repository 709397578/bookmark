<template>
  <div class="settings-view">
    <div class="page-header">
      <div class="header-left">
        <el-button :icon="ArrowLeft" @click="$router.push('/')">返回首页</el-button>
        <h2 class="page-title">
          <el-icon><Setting /></el-icon>
          系统设置
        </h2>
      </div>
    </div>

    <div class="settings-content" v-loading="loading">
      <div class="setting-card">
        <div class="setting-card-header">
          <div class="setting-card-title">用户注册</div>
          <div class="setting-card-desc">控制是否允许新用户注册账号</div>
        </div>
        <div class="setting-card-body">
          <div class="setting-row">
            <div class="setting-info">
              <div class="setting-label">开启注册功能</div>
              <div class="setting-hint">
                {{
                  allowRegistration
                    ? '当前状态：已开启，任何人都可以注册新账号'
                    : '当前状态：已关闭，注册入口将被隐藏'
                }}
              </div>
            </div>
            <el-switch
              v-model="allowRegistration"
              :loading="savingRegistration"
              @change="handleToggleRegistration"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Setting } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { settingAPI } from '@/api'

const router = useRouter()
const authStore = useAuthStore()

const loading = ref(false)
const savingRegistration = ref(false)
const allowRegistration = ref(true)

onMounted(async () => {
  if (!authStore.user && authStore.token) {
    authStore.initAuth()
  }
  await new Promise((resolve) => setTimeout(resolve, 100))

  if (!authStore.isAuthenticated || authStore.user?.role !== 'admin') {
    ElMessage.warning('需要管理员权限')
    router.push('/')
    return
  }
  await loadSettings()
})

const loadSettings = async () => {
  loading.value = true
  try {
    const response: any = await settingAPI.getSettings()
    if (response?.code === 200 || response?.success) {
      const data = response.data
      if (data && typeof data.allowRegistration === 'boolean') {
        allowRegistration.value = data.allowRegistration
      }
    }
  } catch (error) {
    console.error('加载设置失败:', error)
  } finally {
    loading.value = false
  }
}

const handleToggleRegistration = async (val: boolean) => {
  savingRegistration.value = true
  try {
    const response: any = await settingAPI.updateSetting('allowRegistration', val)
    if (response?.code === 200 || response?.success) {
      ElMessage.success(val ? '已开启注册功能' : '已关闭注册功能')
    } else {
      allowRegistration.value = !val
      ElMessage.error(response?.message || '更新失败')
    }
  } catch {
    allowRegistration.value = !val
    ElMessage.error('更新失败')
  } finally {
    savingRegistration.value = false
  }
}
</script>

<style scoped>
.settings-view {
  padding: 1px;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
  margin-top: 4px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #333;
  display: flex;
  align-items: center;
  gap: 6px;
}

.page-title .el-icon {
  color: #667eea;
}

.settings-content {
  flex: 1;
  padding: 12px;
  max-width: 640px;
}

.setting-card {
  background: white;
  border-radius: 10px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
  overflow: hidden;
}

.setting-card-header {
  padding: 16px 20px 12px;
  border-bottom: 1px solid #f0f0f0;
}

.setting-card-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.setting-card-desc {
  font-size: 13px;
  color: #909399;
}

.setting-card-body {
  padding: 16px 20px;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.setting-info {
  flex: 1;
}

.setting-label {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 2px;
}

.setting-hint {
  font-size: 12px;
  color: #909399;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }
}
</style>
