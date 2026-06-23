<template>
  <div class="desktop-layout">
    <el-container>
      <!-- 侧边栏 -->
      <el-aside :width="sidebarCollapsed ? '64px' : '220px'" class="sidebar">
        <div class="sidebar-header">
          <h2 v-if="!sidebarCollapsed">📚 书签</h2>
          <h2 v-else>📚</h2>
          <el-icon class="collapse-btn" @click="sidebarCollapsed = !sidebarCollapsed">
            <Fold />
          </el-icon>
        </div>

        <div class="sidebar-content">
          <!-- 收藏夹列表 -->
          <div class="collections-section">
            <div class="section-header">
              <span v-if="!sidebarCollapsed">收藏夹</span>
            </div>
            <el-skeleton v-if="collectionStore.loading" :rows="3" animated />
            <div v-else class="collection-list">
              <div
                v-for="collection in collectionStore.collections.filter(
                  (c) => c.isPublic || authStore.isAuthenticated,
                )"
                :key="collection.id"
                class="collection-item"
                :class="{ active: collectionStore.currentCollection?.id === collection.id }"
                @click="selectCollection(collection)"
              >
                <span class="collection-icon">{{ collection.icon || '📁' }}</span>
                <span v-if="!sidebarCollapsed" class="collection-name">{{ collection.name }}</span>
                <!-- 显示收藏夹的公开状态 -->
                <el-tooltip
                  v-if="!sidebarCollapsed && !collection.isPublic"
                  content="私有收藏夹（仅自己可见）"
                  placement="right"
                >
                  <el-icon class="private-icon" style="margin-left: 8px; color: #909399">
                    <Lock />
                  </el-icon>
                </el-tooltip>
              </div>
            </div>
          </div>

          <!-- 文件夹列表 -->
          <div
            class="folders-section"
            v-if="collectionStore.currentCollection && !sidebarCollapsed"
          >
            <div class="section-header">
              <span>文件夹</span>
            </div>
            <el-skeleton v-if="collectionStore.loading" :rows="3" animated />
            <div v-else class="folder-list">
              <div
                v-for="folder in collectionStore.folders"
                :key="folder.id"
                class="folder-item"
                :class="{ active: selectedFolder === folder.id }"
                :style="{
                  backgroundColor: folder.color ? `${folder.color}20` : 'transparent',
                  borderLeftColor: folder.color || 'transparent',
                }"
                @click="selectFolder(folder.id)"
              >
                <span v-if="folder.icon" class="folder-custom-icon">{{ folder.icon }}</span>
                <el-icon v-else><Folder /></el-icon>
                <span class="folder-name" :style="{ color: folder.color || 'inherit' }">{{
                  folder.name
                }}</span>
              </div>
              <div
                class="folder-item"
                :class="{ active: selectedFolder === 'none' }"
                @click="selectFolder('none')"
              >
                <el-icon><Document /></el-icon>
                <span class="folder-name">未分类</span>
              </div>
            </div>
          </div>
        </div>
      </el-aside>

      <!-- 主内容区 -->
      <el-container>
        <el-header class="top-bar">
          <!-- 导航栏Breadcrumb 面包屑 -->
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>

            <!-- 当前收藏夹 -->
            <el-breadcrumb-item v-if="collectionStore.currentCollection">
              <span @click="navigateToCollection(collectionStore.currentCollection)">{{
                collectionStore.currentCollection.name
              }}</span>
            </el-breadcrumb-item>

            <!-- 当前文件夹 -->
            <el-breadcrumb-item v-if="selectedFolder && selectedFolder !== 'none'">
              <span @click="navigateToFolder(selectedFolder)">{{
                getFolderName(selectedFolder)
              }}</span>
            </el-breadcrumb-item>

            <!-- 未分类 -->
            <el-breadcrumb-item v-if="selectedFolder === 'none'">未分类</el-breadcrumb-item>

            <!-- 当前页面标识 -->
            <el-breadcrumb-item v-if="$route.name === 'bookmarks'">书签管理</el-breadcrumb-item>
            <el-breadcrumb-item v-if="$route.name === 'bookmark-add'">添加书签</el-breadcrumb-item>
            <el-breadcrumb-item v-if="$route.name === 'bookmark-edit'">编辑书签</el-breadcrumb-item>
            <el-breadcrumb-item v-if="$route.name === 'about'">关于</el-breadcrumb-item>
          </el-breadcrumb>

          <div class="search-box">
            <el-input
              v-model="searchQuery"
              placeholder="搜索书签..."
              clearable
              @keyup.enter="handleSearch"
              @input="handleSearchInput"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </div>
          <div class="actions">
            <!-- 主题切换按钮 -->
            <el-tooltip content="切换主题" placement="bottom">
              <el-button circle @click="toggleDark()">
                <el-icon><Sunny v-if="!isDark" /><Moon v-else /></el-icon>
              </el-button>
            </el-tooltip>

            <!-- 快捷添加书签按钮（已登录时显示） -->
            <el-tooltip v-if="authStore.isAuthenticated" content="添加书签" placement="bottom">
              <el-button type="primary" circle @click="showQuickAddBookmark = true">
                <el-icon><Plus /></el-icon>
              </el-button>
            </el-tooltip>

            <!-- 用户信息下拉菜单（已登录时显示） -->
            <el-dropdown
              v-if="authStore.isAuthenticated"
              trigger="click"
              @command="handleUserCommand"
            >
              <div class="user-avatar-wrapper">
                <el-avatar v-if="authStore.user?.avatar" :size="32" :src="authStore.user.avatar" />
                <el-avatar v-else :size="32" :icon="UserFilled" />
                <span v-if="!sidebarCollapsed" class="user-name-text">{{
                  authStore.user?.name || authStore.user?.email
                }}</span>
                <el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item disabled>
                    <div class="user-info-dropdown">
                      <div class="user-email">{{ authStore.user?.email }}</div>
                      <el-tag
                        size="small"
                        :type="authStore.user?.role === 'admin' ? 'danger' : 'info'"
                      >
                        {{ authStore.user?.role === 'admin' ? '管理员' : '普通用户' }}
                      </el-tag>
                    </div>
                  </el-dropdown-item>

                  <!-- 头像设置 -->
                  <el-dropdown-item divided command="avatar">
                    <el-icon><Picture /></el-icon>
                    设置头像
                  </el-dropdown-item>

                  <!-- 修改密码 -->
                  <el-dropdown-item command="password">
                    <el-icon><Lock /></el-icon>
                    修改密码
                  </el-dropdown-item>

                  <!-- 书签管理 -->
                  <el-dropdown-item divided command="bookmarkManager">
                    <el-icon><Collection /></el-icon>
                    书签管理
                  </el-dropdown-item>

                  <!-- 用户管理（仅管理员） -->
                  <el-dropdown-item
                    v-if="authStore.user?.role === 'admin'"
                    divided
                    command="userManager"
                  >
                    <el-icon><User /></el-icon>
                    用户管理
                  </el-dropdown-item>

                  <!-- 系统设置（仅管理员） -->
                  <el-dropdown-item
                    v-if="authStore.user?.role === 'admin'"
                    divided
                    command="settings"
                  >
                    <el-icon><Setting /></el-icon>
                    系统设置
                  </el-dropdown-item>

                  <el-dropdown-item divided command="logout">
                    <el-icon><SwitchButton /></el-icon>
                    退出登录
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>

            <!-- 登录按钮（未登录时显示） -->
            <el-tooltip v-else content="登录" placement="bottom">
              <el-button circle type="primary" @click="$router.push('/login')">
                <el-icon><User /></el-icon>
              </el-button>
            </el-tooltip>
          </div>
        </el-header>

        <el-main class="content-area">
          <slot></slot>
        </el-main>
      </el-container>
    </el-container>

    <!-- 快捷添加书签对话框 -->
    <el-dialog v-model="showQuickAddBookmark" title="添加书签" width="480px" destroy-on-close>
      <el-form
        ref="quickBookmarkFormRef"
        :model="quickBookmarkForm"
        :rules="quickBookmarkRules"
        label-width="80px"
      >
        <el-form-item label="标题" prop="title">
          <el-input v-model="quickBookmarkForm.title" placeholder="请输入书签标题" />
        </el-form-item>
        <el-form-item label="URL" prop="url">
          <el-input v-model="quickBookmarkForm.url" placeholder="https://example.com" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="quickBookmarkForm.description"
            type="textarea"
            :rows="2"
            placeholder="请输入描述（可选）"
          />
        </el-form-item>
        <el-form-item label="收藏夹" prop="collectionId">
          <el-select
            v-model="quickBookmarkForm.collectionId"
            placeholder="请选择收藏夹"
            style="width: 100%"
            @change="onQuickCollectionChange"
          >
            <el-option
              v-for="c in collectionStore.collections.filter(
                (c) => c.isPublic || authStore.isAuthenticated,
              )"
              :key="c.id"
              :label="`${c.icon || '📁'} ${c.name}`"
              :value="c.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="文件夹">
          <el-select
            v-model="quickBookmarkForm.folderId"
            placeholder="未分类"
            clearable
            style="width: 100%"
            :disabled="!quickBookmarkForm.collectionId"
          >
            <el-option v-for="f in quickFolderList" :key="f.id" :label="f.name" :value="f.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showQuickAddBookmark = false">取消</el-button>
        <el-button type="primary" @click="handleQuickAddBookmark" :loading="quickAddLoading">
          保存
        </el-button>
      </template>
    </el-dialog>

    <!-- 头像设置对话框 -->
    <el-dialog v-model="showAvatarDialog" title="设置头像" width="400px">
      <div class="avatar-dialog-content">
        <div class="avatar-preview">
          <el-avatar v-if="authStore.user?.avatar" :size="96" :src="authStore.user.avatar" />
          <el-avatar v-else :size="96" :icon="UserFilled" />
        </div>
        <el-upload
          ref="avatarUploadRef"
          class="avatar-uploader"
          :auto-upload="false"
          :show-file-list="false"
          accept="image/*"
          :on-change="handleAvatarChange"
        >
          <el-button type="primary">
            <el-icon><Upload /></el-icon>
            选择图片
          </el-button>
        </el-upload>
        <div class="avatar-tips">支持 JPG、PNG 格式，文件大小不超过 2MB</div>
        <div v-if="authStore.user?.avatar" class="avatar-actions">
          <el-button type="danger" plain size="small" @click="handleDeleteAvatar">
            删除头像
          </el-button>
        </div>
      </div>
      <template #footer>
        <el-button @click="showAvatarDialog = false">取消</el-button>
        <el-button type="primary" :loading="avatarUploading" @click="handleUploadAvatar">
          保存
        </el-button>
      </template>
    </el-dialog>

    <!-- 修改密码对话框 -->
    <el-dialog
      v-model="showPasswordDialog"
      title="修改密码"
      width="420px"
      destroy-on-close
      @closed="resetPasswordForm"
    >
      <el-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        label-width="90px"
      >
        <el-form-item label="旧密码" prop="oldPassword">
          <el-input
            v-model="passwordForm.oldPassword"
            type="password"
            show-password
            placeholder="请输入当前密码"
          />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input
            v-model="passwordForm.newPassword"
            type="password"
            show-password
            placeholder="至少8位，需包含字母和数字"
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="passwordForm.confirmPassword"
            type="password"
            show-password
            placeholder="请再次输入新密码"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPasswordDialog = false">取消</el-button>
        <el-button type="primary" :loading="passwordSubmitting" @click="handleChangePassword">
          确认修改
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Fold,
  Plus,
  UserFilled,
  Folder,
  FolderAdd,
  Document,
  Search,
  Sunny,
  Moon,
  SwitchButton,
  User,
  ArrowDown,
  Collection,
  Picture,
  Upload,
  Setting,
  Lock,
} from '@element-plus/icons-vue'
import { useDark, useToggle } from '@vueuse/core'
import { useAuthStore } from '@/stores/auth'
import { useCollectionStore } from '@/stores/collection'
import { authAPI } from '@/api'

const isDark = useDark()
const toggleDark = useToggle(isDark)

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const collectionStore = useCollectionStore()

const sidebarCollapsed = ref(false)
const searchQuery = ref('')
const selectedFolder = ref<string | null>(null)
const searchTimeout = ref<number | null>(null)

// 快捷添加书签
const showQuickAddBookmark = ref(false)
const quickAddLoading = ref(false)
const quickBookmarkFormRef = ref<any>(null)

// 头像相关
const showAvatarDialog = ref(false)
const avatarUploading = ref(false)
const avatarFile = ref<File | null>(null)
const quickBookmarkForm = reactive({
  title: '',
  url: '',
  description: '',
  collectionId: '' as string,
  folderId: '' as string,
})
const quickBookmarkRules = {
  title: [{ required: true, message: '请输入书签标题', trigger: 'blur' }],
  url: [{ required: true, message: '请输入网址', trigger: 'blur' }],
  collectionId: [{ required: true, message: '请选择收藏夹', trigger: 'change' }],
}
const quickFolderList = ref<any[]>([])

// 收藏夹切换时加载文件夹
const onQuickCollectionChange = async (collectionId: string) => {
  quickBookmarkForm.folderId = ''
  if (collectionId) {
    const folders = await collectionStore.fetchFolders(collectionId)
    quickFolderList.value = collectionStore.folders
  } else {
    quickFolderList.value = []
  }
}

// 重置表单
const resetQuickBookmarkForm = () => {
  quickBookmarkForm.title = ''
  quickBookmarkForm.url = ''
  quickBookmarkForm.description = ''
  quickBookmarkForm.collectionId = ''
  quickBookmarkForm.folderId = ''
  quickFolderList.value = []
}

// 处理搜索输入（带防抖）
const handleSearchInput = () => {
  if (searchTimeout.value) {
    clearTimeout(searchTimeout.value)
  }
  searchTimeout.value = window.setTimeout(() => {
    if (searchQuery.value.trim()) {
      handleSearch()
    } else {
      // 当搜索内容为空时，清除搜索结果
      window.dispatchEvent(
        new CustomEvent('clearSearchResults', {
          detail: {},
        }),
      )
    }
  }, 300)
}

// 对话框状态
const showCreateFolder = ref(false)

const newFolder = reactive({
  name: '',
})

// 初始化
onMounted(async () => {
  // 确保认证状态已初始化
  if (!authStore.user && authStore.token) {
    authStore.initAuth()
  }

  // 等待一小段时间确保状态完全更新
  await new Promise((resolve) => setTimeout(resolve, 100))

  // 根据登录状态决定是否只获取公开收藏夹
  const publicOnly = !authStore.isAuthenticated
  await collectionStore.fetchCollections(publicOnly)
})

// 选择收藏夹
const selectCollection = async (collection: any) => {
  // 检查用户是否有权限访问该收藏夹
  if (!collection.isPublic && !authStore.isAuthenticated) {
    ElMessage.warning('请先登录后再访问私有收藏夹')
    return
  }

  // 清除搜索框内容和结果
  searchQuery.value = ''
  window.dispatchEvent(
    new CustomEvent('clearSearchResults', {
      detail: {},
    }),
  )

  collectionStore.setCurrentCollection(collection)
  selectedFolder.value = null
  await collectionStore.fetchFolders(collection.id)
  await collectionStore.fetchBookmarks({ collectionId: collection.id })
}

// 选择文件夹
const selectFolder = async (folderId: string | null) => {
  // 清除搜索框内容和结果
  searchQuery.value = ''
  window.dispatchEvent(
    new CustomEvent('clearSearchResults', {
      detail: {},
    }),
  )

  selectedFolder.value = folderId
  if (collectionStore.currentCollection) {
    await collectionStore.fetchBookmarks({
      collectionId: collectionStore.currentCollection.id,
      folderId: folderId || undefined,
    })
  }
}

// 导航到书签管理页面
const openBookmarkManager = () => {
  router.push({ name: 'bookmarks' })
}

// 用户操作命令处理
const handleUserCommand = (command: string) => {
  if (command === 'logout') {
    handleLogout()
  } else if (command === 'bookmarkManager') {
    openBookmarkManager()
  } else if (command === 'userManager') {
    router.push({ name: 'user-manage' })
  } else if (command === 'settings') {
    router.push({ name: 'settings' })
  } else if (command === 'avatar') {
    showAvatarDialog.value = true
  } else if (command === 'password') {
    showPasswordDialog.value = true
  }
}

// 头像上传
const handleAvatarChange = (file: any) => {
  const rawFile = file.raw as File
  if (!rawFile) return
  if (rawFile.size > 2 * 1024 * 1024) {
    ElMessage.warning('头像文件不能超过2MB')
    return
  }
  avatarFile.value = rawFile
}

const handleUploadAvatar = async () => {
  if (!avatarFile.value) {
    ElMessage.warning('请先选择图片')
    return
  }
  avatarUploading.value = true
  try {
    const response: any = await authAPI.uploadAvatar(avatarFile.value)
    if (response?.code === 200 || response?.success) {
      const avatarUrl = response.data?.avatar
      if (authStore.user) {
        authStore.user.avatar = avatarUrl
        localStorage.setItem('user', JSON.stringify(authStore.user))
      }
      ElMessage.success('头像上传成功')
      showAvatarDialog.value = false
      avatarFile.value = null
    } else {
      ElMessage.error(response?.message || '上传失败')
    }
  } catch (error) {
    ElMessage.error('上传失败')
  } finally {
    avatarUploading.value = false
  }
}

const handleDeleteAvatar = async () => {
  try {
    const response: any = await authAPI.deleteAvatar()
    if (response?.code === 200 || response?.success) {
      if (authStore.user) {
        authStore.user.avatar = ''
        localStorage.setItem('user', JSON.stringify(authStore.user))
      }
      ElMessage.success('头像已删除')
    }
  } catch {
    ElMessage.error('删除失败')
  }
}

// 修改密码相关
const showPasswordDialog = ref(false)
const passwordSubmitting = ref(false)
const passwordFormRef = ref<any>(null)
const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

// 密码强度校验：至少8位，包含字母和数字
const validatePasswordStrength = (_rule: any, value: string, callback: any) => {
  if (!value) {
    callback(new Error('请输入新密码'))
  } else if (value.length < 8) {
    callback(new Error('密码长度至少8位'))
  } else if (!/[a-zA-Z]/.test(value)) {
    callback(new Error('密码必须包含至少一个字母'))
  } else if (!/[0-9]/.test(value)) {
    callback(new Error('密码必须包含至少一个数字'))
  } else {
    callback()
  }
}

const validateConfirmPassword = (_rule: any, value: string, callback: any) => {
  if (!value) {
    callback(new Error('请再次输入新密码'))
  } else if (value !== passwordForm.newPassword) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const passwordRules = {
  oldPassword: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  newPassword: [{ required: true, validator: validatePasswordStrength, trigger: 'blur' }],
  confirmPassword: [{ required: true, validator: validateConfirmPassword, trigger: 'blur' }],
}

const resetPasswordForm = () => {
  passwordForm.oldPassword = ''
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
  passwordFormRef.value?.clearValidate()
}

const handleChangePassword = async () => {
  try {
    await passwordFormRef.value?.validate()
  } catch {
    return
  }
  passwordSubmitting.value = true
  try {
    const response: any = await authAPI.changePassword({
      oldPassword: passwordForm.oldPassword,
      newPassword: passwordForm.newPassword,
    })
    if (response?.code === 200 || response?.success) {
      ElMessage.success('密码修改成功')
      showPasswordDialog.value = false
    } else {
      ElMessage.error(response?.message || '修改失败')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '修改失败')
  } finally {
    passwordSubmitting.value = false
  }
}

// 搜索书签
const handleSearch = async () => {
  if (!searchQuery.value) {
    return
  }

  try {
    const results = await collectionStore.searchBookmarks(searchQuery.value)
    if (results && results.length > 0) {
      // 保存搜索结果到sessionStorage
      window.sessionStorage.setItem('isSearchMode', 'true')
      window.sessionStorage.setItem('searchQuery', searchQuery.value)
      window.sessionStorage.setItem('searchResults', JSON.stringify(results))

      // 触发全局事件通知HomeView显示搜索结果
      window.dispatchEvent(
        new CustomEvent('searchResultsUpdated', {
          detail: {
            query: searchQuery.value,
            results: results,
          },
        }),
      )

      // ElMessage.success(`找到 ${results.length} 个结果`)
    } else {
      ElMessage.info(`未找到匹配的书签`)
    }
  } catch (error) {
    console.error('搜索失败:', error)
    ElMessage.error('搜索失败')
  }
}

// 退出登录
const handleLogout = () => {
  ElMessageBox.confirm('确定要退出登录吗？', '确认退出', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
    .then(() => {
      authStore.logout()
      router.push('/')
      ElMessage.success('已退出登录')
    })
    .catch(() => {
      // 用户取消操作，不做处理
    })
}

// 快捷添加书签
const handleQuickAddBookmark = async () => {
  try {
    await quickBookmarkFormRef.value?.validate()
  } catch {
    return
  }

  quickAddLoading.value = true
  try {
    const result = await collectionStore.createBookmark({
      title: quickBookmarkForm.title,
      url: quickBookmarkForm.url,
      description: quickBookmarkForm.description,
      collectionId: quickBookmarkForm.collectionId,
      folderId: quickBookmarkForm.folderId || undefined,
    })

    if (result.success) {
      ElMessage.success('添加成功')
      showQuickAddBookmark.value = false
      resetQuickBookmarkForm()
    } else {
      ElMessage.error(result.message || '添加失败')
    }
  } catch (error) {
    ElMessage.error('添加失败')
  } finally {
    quickAddLoading.value = false
  }
}

// 导航到收藏夹
const navigateToCollection = (collection: any) => {
  selectCollection(collection)
}

// 导航到文件夹
const navigateToFolder = async (folderId: string) => {
  if (!collectionStore.currentCollection) return

  // 选择文件夹
  await selectFolder(folderId)
}

// 获取文件夹名称
const getFolderName = (folderId: string | null) => {
  if (!folderId || folderId === 'none') return '未分类'

  const folder = collectionStore.folders.find((f) => f.id === folderId)
  return folder ? folder.name : '未知文件夹'
}

// 监听路由变化
watch(
  () => route.name,
  (newName) => {
    // 根据路由名称执行相应的操作
    if (newName === 'home') {
      // 如果是首页，确保加载默认收藏夹
      if (collectionStore.collections.length > 0 && !collectionStore.currentCollection) {
        selectCollection(collectionStore.collections[0])
      }
    }
  },
  { immediate: true },
)
</script>

<style scoped>
.desktop-layout {
  height: 100vh;
}

.el-container {
  height: 100%;
}

.sidebar {
  background: white;
  border-right: 1px solid #e8e8e8;
  display: flex;
  flex-direction: column;
  transition: all 0.3s ease;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.05);
}

.sidebar-header {
  padding: 14px;
  border-bottom: 1px solid #e8e8e8;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.sidebar-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.collapse-btn {
  cursor: pointer;
  font-size: 18px;
  color: white;
  transition: transform 0.3s;
}

.collapse-btn:hover {
  transform: scale(1.1);
}

.sidebar-content {
  flex: 1;
  overflow-y: auto;
  padding: 6px;
}

.login-prompt {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 24px 16px;
  margin-bottom: 20px;
  text-align: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  color: white;
}

.login-prompt p {
  margin: 12px 0;
  font-size: 14px;
  font-weight: 500;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  border-radius: 8px;
  margin-bottom: 20px;
  color: white;
  box-shadow: 0 2px 8px rgba(245, 87, 108, 0.3);
}

.user-details {
  flex: 1;
}

.user-name {
  font-weight: 600;
  font-size: 14px;
}

.user-role {
  font-size: 12px;
  opacity: 0.9;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
  font-weight: 600;
  color: #333;
  font-size: 13px;
  padding-top: 6px;
}

.collections-section {
  margin-bottom: 8px;
}

.folders-section {
  margin-top: 8px;
}

.collection-list,
.folder-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.collection-item,
.folder-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

.collection-item:hover,
.folder-item:hover {
  background: #f0f0f0;
  border-color: #d9d9d9;
  transform: translateX(4px);
}

.collection-item.active,
.folder-item.active {
  background: linear-gradient(135deg, #e6f7ff 0%, #bae7ff 100%);
  color: #1890ff;
  border-color: #1890ff;
  font-weight: 500;
}

.folder-item {
  border-left: 4px solid transparent;
  transition: all 0.2s ease;
}

.folder-item:hover {
  border-left-color: rgba(0, 0, 0, 0.2);
}

.folder-item.active {
  border-left-color: #1890ff;
}

.collection-icon {
  font-size: 18px;
}

.folder-item .el-icon {
  font-size: 16px;
}

.folder-custom-icon {
  font-size: 16px;
  width: 16px;
  text-align: center;
  flex-shrink: 0;
}

.collection-name,
.folder-name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 13px;
  line-height: 1.2;
}

.top-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 24px;
  background: white;
  border-bottom: 1px solid #e8e8e8;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.search-box {
  flex: 1;
  max-width: 500px;
  margin: 0 16px;
}

.search-box .el-input {
  border-radius: 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  transition: all 0.3s;
}

.search-box .el-input:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
}

.search-box .el-input.is-focus {
  box-shadow: 0 4px 16px rgba(64, 158, 255, 0.2);
}

.search-box .el-input__inner {
  border-radius: 20px;
  padding-left: 40px;
  font-size: 14px;
  height: 40px;
}

.search-box .el-input__prefix {
  left: 12px;
  color: #909399;
}

.search-box .el-input__suffix {
  right: 12px;
}

.search-box .el-input__suffix .el-input__clear {
  color: #909399;
  font-size: 16px;
  transition: all 0.2s;
}

.search-box .el-input__suffix .el-input__clear:hover {
  color: #409eff;
  transform: scale(1.2);
}

.content-area {
  padding: 1px 5px;
  background: #f5f5f5;
}

.actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.about-info {
  color: #606266;
  font-size: 14px;
  line-height: 1.8;
}

.about-info p {
  margin: 4px 0;
}

.user-avatar-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.user-name-text {
  font-size: 14px;
  font-weight: 500;
}

.user-info-dropdown {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.user-email {
  font-size: 14px;
  font-weight: 500;
}

.avatar-dialog-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 10px 0;
}

.avatar-preview {
  margin-bottom: 8px;
}

.avatar-tips {
  font-size: 12px;
  color: #909399;
}

.avatar-actions {
  margin-top: 4px;
}

.top-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  background: white;
  border-bottom: 1px solid #e8e8e8;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.search-box {
  flex: 1;
  max-width: 500px;
  margin: 0 12px;
  /* 可选：如果希望搜索框在中间居中，可以保留 flex: 1；
     如果希望搜索框紧跟面包屑，而按钮靠右，目前的结构配合 margin-left: auto 即可 */
}

/* 修改这里：确保 actions 区域靠右对齐 */
.actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto; /* 关键属性：将元素推到最右侧 */
}
</style>
