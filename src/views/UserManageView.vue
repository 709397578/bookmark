<template>
  <div class="user-manage-view">
    <div class="page-header">
      <div class="header-left">
        <el-button :icon="ArrowLeft" @click="$router.push('/')">返回首页</el-button>
        <h2 class="page-title">
          <el-icon><User /></el-icon>
          用户管理
        </h2>
      </div>
    </div>

    <div class="filter-section">
      <el-input v-model="searchQuery" placeholder="搜索用户..." clearable style="width: 300px">
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <div class="filter-stats">
        共 <strong>{{ filteredUsers.length }}</strong> 个用户
      </div>
    </div>

    <div class="table-wrapper" v-loading="loading">
      <el-table
        v-if="filteredUsers.length > 0"
        :data="filteredUsers"
        stripe
        style="width: 100%"
        height="100%"
        :header-cell-style="{ background: '#fafafa', color: '#333', fontWeight: 600 }"
      >
        <el-table-column label="用户" min-width="200">
          <template #default="{ row }">
            <div class="user-cell">
              <el-avatar :size="36" :icon="UserFilled" />
              <div class="user-cell-info">
                <div class="user-cell-name">{{ row.name || '未设置名称' }}</div>
                <div class="user-cell-email">{{ row.email }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="120" align="center">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'danger' : 'info'" size="small">
              {{ row.role === 'admin' ? '管理员' : '普通用户' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="收藏夹" width="100" align="center">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.collectionCount ?? 0 }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="书签" width="100" align="center">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.bookmarkCount ?? 0 }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="注册时间" width="170">
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right" align="center">
          <template #default="{ row }">
            <el-button size="small" type="primary" plain :icon="Edit" @click="handleEditUser(row)">
              编辑
            </el-button>
            <el-button
              size="small"
              type="danger"
              plain
              :icon="Delete"
              :disabled="row.id === authStore.user?.id"
              @click="handleDeleteUser(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无用户数据" />
    </div>

    <!-- 编辑用户对话框 -->
    <el-dialog v-model="showEditDialog" title="编辑用户" width="400px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="邮箱">
          <el-input :model-value="editForm.email" disabled />
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="editForm.name" placeholder="请输入用户名称" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="editForm.role" style="width: 100%">
            <el-option label="普通用户" value="user" />
            <el-option label="管理员" value="admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveUser">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, User, Search, Edit, Delete, UserFilled } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { authAPI } from '@/api'

const router = useRouter()
const authStore = useAuthStore()

const users = ref<any[]>([])
const loading = ref(false)
const searchQuery = ref('')
const showEditDialog = ref(false)

const editForm = reactive({
  id: '',
  email: '',
  name: '',
  role: '',
})

const filteredUsers = computed(() => {
  if (!searchQuery.value) return users.value
  const q = searchQuery.value.toLowerCase()
  return users.value.filter(
    (u) => u.name?.toLowerCase().includes(q) || u.email?.toLowerCase().includes(q),
  )
})

onMounted(async () => {
  // 确保认证状态已初始化
  if (!authStore.user && authStore.token) {
    authStore.initAuth()
  }

  // 等待状态更新
  await new Promise((resolve) => setTimeout(resolve, 100))

  if (!authStore.isAuthenticated || authStore.user?.role !== 'admin') {
    ElMessage.warning('需要管理员权限')
    router.push('/')
    return
  }
  await loadUsers()
})

const loadUsers = async () => {
  loading.value = true
  try {
    const response: any = await authAPI.listUsers()
    if (response?.code === 200 || response?.success) {
      users.value = response.data || []
    }
  } catch (error) {
    console.error('加载用户列表失败:', error)
    ElMessage.error('加载用户列表失败')
  } finally {
    loading.value = false
  }
}

const handleEditUser = (user: any) => {
  editForm.id = user.id
  editForm.email = user.email
  editForm.name = user.name || ''
  editForm.role = user.role
  showEditDialog.value = true
}

const handleSaveUser = async () => {
  try {
    const response: any = await authAPI.updateUser(editForm.id, {
      name: editForm.name,
      role: editForm.role,
    })
    if (response?.code === 200 || response?.success) {
      ElMessage.success('更新成功')
      showEditDialog.value = false
      await loadUsers()
    } else {
      ElMessage.error(response?.message || '更新失败')
    }
  } catch (error) {
    ElMessage.error('更新失败')
  }
}

const handleDeleteUser = async (user: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户"${user.name || user.email}"吗？该用户的收藏夹和书签将一并删除。`,
      '确认删除',
      { confirmButtonText: '确定删除', cancelButtonText: '取消', type: 'warning' },
    )
    const response: any = await authAPI.deleteUser(user.id)
    if (response?.code === 200 || response?.success) {
      ElMessage.success('删除成功')
      await loadUsers()
    } else {
      ElMessage.error(response?.message || '删除失败')
    }
  } catch {
    // 用户取消
  }
}

const formatDate = (timeStr: string) => {
  if (!timeStr) return ''
  return new Date(timeStr).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<style scoped>
.user-manage-view {
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

.filter-section {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
  flex-wrap: wrap;
  margin-top: 4px;
}

.filter-stats {
  margin-left: auto;
  font-size: 14px;
  color: #666;
}

.filter-stats strong {
  color: #667eea;
  font-size: 16px;
}

.table-wrapper {
  flex: 1;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
  overflow: auto;
  padding: 12px;
  margin-top: 4px;
  max-height: calc(100vh - 200px);
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-cell-info {
  min-width: 0;
}

.user-cell-name {
  font-weight: 500;
  color: #333;
  font-size: 14px;
}

.user-cell-email {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}

@media (max-width: 1024px) {
  .page-header {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }

  .filter-section {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-stats {
    margin-left: 0;
    text-align: center;
  }
}
</style>
