<template>
  <div class="app-layout">
    <!-- 侧边栏遮罩 -->
    <transition name="fade">
      <div v-if="!sidebarCollapsed" class="sidebar-overlay" @click="sidebarCollapsed = true" />
    </transition>

    <!-- 侧边栏 -->
    <transition name="slide">
      <aside v-if="!sidebarCollapsed" class="sidebar">
        <div class="sidebar-header">
          <h2>📚 书签</h2>
          <van-icon name="cross" class="close-btn" @click="sidebarCollapsed = true" />
        </div>

        <div class="sidebar-content">
          <!-- 未登录提示 -->
          <div v-if="!authStore.isAuthenticated" class="login-prompt">
            <div class="login-avatar">
              <van-icon name="user-circle-o" size="48" color="rgba(255,255,255,0.9)" />
            </div>
            <p>登录后享受更多功能</p>
            <van-button
              round
              size="small"
              type="primary"
              block
              class="login-btn"
              @click="$router.push('/login')"
            >
              立即登录
            </van-button>
          </div>

          <!-- 用户信息 -->
          <div class="user-info" v-else-if="authStore.user">
            <div class="user-avatar">
              <van-icon name="user-circle-o" size="28" />
            </div>
            <div class="user-details">
              <div class="user-name">{{ authStore.user.name || authStore.user.email }}</div>
              <div class="user-role">{{ authStore.user.role === 'admin' ? '管理员' : '用户' }}</div>
            </div>
          </div>

          <!-- 账户操作 -->
          <div class="menu-section" v-if="authStore.isAuthenticated">
            <div class="section-label">账户</div>
            <div class="menu-list">
              <div class="menu-item" @click="showPasswordDialog = true">
                <span class="folder-custom-icon">🔑</span>
                <span class="menu-text">修改密码</span>
              </div>
            </div>
          </div>

          <!-- 收藏夹列表 -->
          <div class="menu-section">
            <div class="section-label">收藏夹</div>
            <div class="menu-list">
              <div
                v-for="collection in collectionStore.collections"
                :key="collection.id"
                class="menu-item"
                :class="{ active: collectionStore.currentCollection?.id === collection.id }"
                @click="selectCollection(collection)"
              >
                <span class="menu-icon">{{ collection.icon || '📁' }}</span>
                <span class="menu-text">{{ collection.name }}</span>
                <van-icon
                  v-if="collectionStore.currentCollection?.id === collection.id"
                  name="success"
                  class="check-icon"
                />
              </div>
              <div
                v-if="authStore.isAuthenticated"
                class="menu-item add-item"
                @click="showCreateCollection = true"
              >
                <span class="menu-icon">+</span>
                <span class="menu-text">新建收藏夹</span>
              </div>
            </div>
          </div>

          <!-- 文件夹列表 -->
          <div class="menu-section" v-if="collectionStore.currentCollection">
            <div class="section-label">文件夹</div>
            <div class="menu-list">
              <div
                v-for="folder in collectionStore.folders"
                :key="folder.id"
                class="menu-item"
                :class="{ active: selectedFolder === folder.id }"
                :style="{
                  borderLeftColor: folder.color || 'transparent',
                }"
                @click="selectFolder(folder.id)"
              >
                <span class="folder-custom-icon">{{ folder.icon || '📂' }}</span>
                <span class="menu-text" :style="{ color: folder.color || 'inherit' }">{{
                  folder.name
                }}</span>
              </div>
              <div
                class="menu-item"
                :class="{ active: selectedFolder === 'none' }"
                @click="selectFolder('none')"
              >
                <span class="folder-custom-icon">📄</span>
                <span class="menu-text">未分类</span>
              </div>
              <div
                v-if="authStore.isAuthenticated"
                class="menu-item add-item"
                @click="showCreateFolder = true"
              >
                <span class="menu-icon">+</span>
                <span class="menu-text">新建文件夹</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 侧边栏底部 -->
        <div class="sidebar-footer">
          <van-button
            v-if="authStore.isAuthenticated"
            block
            round
            plain
            type="default"
            size="small"
            @click="handleLogout"
          >
            退出登录
          </van-button>
          <van-button
            v-else
            block
            round
            type="primary"
            size="small"
            @click="$router.push('/login')"
          >
            登录 / 注册
          </van-button>
        </div>
      </aside>
    </transition>

    <!-- 主内容区 -->
    <main class="main-content">
      <!-- 顶部导航 -->
      <header class="top-bar">
        <van-icon name="bars" class="menu-btn" @click="sidebarCollapsed = false" />
        <van-search
          v-model="searchQuery"
          placeholder="搜索书签..."
          shape="round"
          background="transparent"
          @search="handleSearch"
        />
        <van-icon
          v-if="authStore.isAuthenticated"
          name="add-o"
          class="add-btn"
          @click="showCreateBookmark = true"
        />
      </header>

      <!-- 内容区域 -->
      <div class="content-area">
        <slot></slot>
      </div>
    </main>

    <!-- 创建收藏夹对话框 -->
    <van-dialog
      v-model:show="showCreateCollection"
      title="创建收藏夹"
      show-cancel-button
      @confirm="handleCreateCollection"
    >
      <van-form>
        <van-field
          v-model="newCollection.name"
          label="名称"
          placeholder="请输入收藏夹名称"
          required
        />
        <van-field
          v-model="newCollection.description"
          label="描述"
          placeholder="请输入描述（可选）"
          type="textarea"
        />
        <van-field
          v-model="newCollection.icon"
          label="图标"
          placeholder="请输入emoji图标（可选）"
        />
      </van-form>
    </van-dialog>

    <!-- 创建文件夹对话框 -->
    <van-dialog
      v-model:show="showCreateFolder"
      title="创建文件夹"
      show-cancel-button
      @confirm="handleCreateFolder"
    >
      <van-form>
        <van-field v-model="newFolder.name" label="名称" placeholder="请输入文件夹名称" required />
      </van-form>
    </van-dialog>

    <!-- 创建/编辑书签对话框 -->
    <van-dialog
      v-model:show="showCreateBookmark"
      :title="editingBookmark ? '编辑书签' : '添加书签'"
      show-cancel-button
      class="bookmark-dialog"
      @confirm="handleBookmarkConfirm"
      @cancel="resetBookmarkForm"
    >
      <van-form ref="bookmarkFormRef" :validate-trigger="['onBlur', 'onChange']">
        <van-field
          v-model="bookmarkForm.title"
          placeholder="请输入书签标题"
          required
          :rules="[{ required: true, message: '请输入书签标题' }]"
        />
        <van-field
          v-model="bookmarkForm.url"
          placeholder="https://example.com"
          required
          :rules="[{ required: true, message: '请输入网址' }]"
        />
        <van-field
          v-model="bookmarkForm.description"
          placeholder="请输入描述（可选）"
          type="textarea"
          rows="2"
          autosize
        />
        <van-field
          :model-value="currentCollectionName"
          placeholder="选择收藏夹"
          readonly
          is-link
          required
          :rules="[{ validator: () => !!bookmarkForm.collectionId, message: '请选择收藏夹' }]"
          @click="showCollectionPicker = true"
        />
        <van-field
          :model-value="currentFolderName"
          placeholder="选择文件夹（可选）"
          readonly
          is-link
          @click="showFolderPicker = true"
          :disabled="!bookmarkForm.collectionId"
        />
      </van-form>
    </van-dialog>

    <!-- 收藏夹选择器 -->
    <van-popup v-model:show="showCollectionPicker" position="bottom" round>
      <van-picker
        :columns="collectionColumns"
        @confirm="onCollectionPick"
        @cancel="showCollectionPicker = false"
      />
    </van-popup>

    <!-- 文件夹选择器 -->
    <van-popup v-model:show="showFolderPicker" position="bottom" round>
      <van-picker
        :columns="folderColumns"
        @confirm="onFolderPick"
        @cancel="showFolderPicker = false"
      />
    </van-popup>

    <!-- 修改密码对话框 -->
    <van-dialog
      v-model:show="showPasswordDialog"
      title="修改密码"
      show-cancel-button
      :before-close="handlePasswordBeforeClose"
      @closed="resetPasswordForm"
    >
      <van-form ref="passwordFormRef">
        <van-field
          v-model="passwordForm.oldPassword"
          type="password"
          label="旧密码"
          placeholder="请输入当前密码"
          required
          :rules="[{ required: true, message: '请输入当前密码' }]"
        />
        <van-field
          v-model="passwordForm.newPassword"
          type="password"
          label="新密码"
          placeholder="至少8位，需含字母和数字"
          required
          :rules="[
            { required: true, message: '请输入新密码' },
            { validator: validatePasswordStrength, message: '密码至少8位，需含字母和数字' },
          ]"
        />
        <van-field
          v-model="passwordForm.confirmPassword"
          type="password"
          label="确认密码"
          placeholder="请再次输入新密码"
          required
          :rules="[
            { required: true, message: '请确认新密码' },
            { validator: validateConfirmPassword, message: '两次输入的密码不一致' },
          ]"
        />
      </van-form>
    </van-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { showToast as showVantToast, showDialog as showVantDialog } from 'vant'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { useCollectionStore } from '@/stores/collection'
import { authAPI } from '@/api'
import { isMobile } from '@/utils/device'

const router = useRouter()
const authStore = useAuthStore()
const collectionStore = useCollectionStore()

const deviceType = computed(() => (isMobile() ? 'mobile' : 'desktop'))

// 侧边栏状态（true = 收起/隐藏）
const sidebarCollapsed = ref(true)
const searchQuery = ref('')
const selectedFolder = ref<string | null>(null)

// 对话框状态
const showCreateCollection = ref(false)
const showCreateFolder = ref(false)
const showCreateBookmark = ref(false)
const editingBookmark = ref(false)

// 表单数据
const newCollection = reactive({ name: '', description: '', icon: '' })
const newFolder = reactive({ name: '' })
const bookmarkForm = reactive({
  title: '',
  url: '',
  description: '',
  icon: '',
  collectionId: undefined as string | undefined,
  folderId: undefined as string | undefined,
})

// 收藏夹/文件夹选择器状态
const showCollectionPicker = ref(false)
const showFolderPicker = ref(false)

// 表单引用
const bookmarkFormRef = ref<any>(null)

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
const validatePasswordStrength = (val: string) => {
  if (val.length < 8) return false
  if (!/[a-zA-Z]/.test(val)) return false
  if (!/[0-9]/.test(val)) return false
  return true
}

const validateConfirmPassword = (val: string) => val === passwordForm.newPassword

const resetPasswordForm = () => {
  passwordForm.oldPassword = ''
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
  passwordFormRef.value?.resetValidation?.()
}

// before-close：点击确认时拦截，校验并提交后再关闭
const handlePasswordBeforeClose = async (action: string) => {
  if (action !== 'confirm') return true
  try {
    await passwordFormRef.value?.validate()
  } catch {
    return false
  }
  passwordSubmitting.value = true
  try {
    const response: any = await authAPI.changePassword({
      oldPassword: passwordForm.oldPassword,
      newPassword: passwordForm.newPassword,
    })
    if (response?.code === 200 || response?.success) {
      showToast('密码修改成功')
      return true
    }
    showToast(response?.message || '修改失败')
    return false
  } catch (error: any) {
    showToast(error.response?.data?.message || '修改失败')
    return false
  } finally {
    passwordSubmitting.value = false
  }
}

// 统一的 Toast 方法
const showToast = (message: string) => {
  if (deviceType.value === 'mobile') {
    showVantToast(message)
  } else {
    ElMessage.success(message)
  }
}

// 初始化
onMounted(async () => {
  if (!authStore.user && authStore.token) {
    authStore.initAuth()
  }
  await collectionStore.fetchCollections()
})

// 搜索内容清空时通知HomeView
watch(searchQuery, (val) => {
  if (!val) {
    window.dispatchEvent(new CustomEvent('clearSearchResults'))
  }
})

// 收藏夹选择器
const currentCollectionName = computed(() => {
  if (!bookmarkForm.collectionId) return ''
  const c = collectionStore.collections.find((c: any) => c.id === bookmarkForm.collectionId)
  return c ? `${c.icon || ''} ${c.name}` : ''
})

const collectionColumns = computed(() =>
  collectionStore.collections
    .filter((c: any) => c.isPublic || authStore.isAuthenticated)
    .map((c: any) => ({ text: `${c.icon || '📁'} ${c.name}`, value: c.id })),
)

const onCollectionPick = ({ selectedOptions }: any) => {
  const selected = selectedOptions[0]
  if (selected) {
    bookmarkForm.collectionId = selected.value
    bookmarkForm.folderId = undefined
    collectionStore.fetchFolders(selected.value)
  }
  showCollectionPicker.value = false
}

// 文件夹选择器
const currentFolderName = computed(() => {
  if (!bookmarkForm.folderId) return ''
  const f = collectionStore.folders.find((f: any) => f.id === bookmarkForm.folderId)
  return f ? f.name : ''
})

const folderColumns = computed(() => [
  { text: '未分类', value: '' },
  ...collectionStore.folders.map((f: any) => ({ text: f.name, value: f.id })),
])

const onFolderPick = ({ selectedOptions }: any) => {
  const selected = selectedOptions[0]
  bookmarkForm.folderId = selected?.value || undefined
  showFolderPicker.value = false
}

// 搜索
const handleSearch = async () => {
  if (!searchQuery.value) return
  const results = await collectionStore.searchBookmarks(searchQuery.value)
  if (results && results.length > 0) {
    window.sessionStorage.setItem('isSearchMode', 'true')
    window.sessionStorage.setItem('searchQuery', searchQuery.value)
    window.sessionStorage.setItem('searchResults', JSON.stringify(results))

    window.dispatchEvent(
      new CustomEvent('searchResultsUpdated', {
        detail: {
          query: searchQuery.value,
          results: results,
        },
      }),
    )

    showToast(`找到 ${results.length} 个结果`)
  } else {
    showToast('未找到匹配的书签')
  }
}

// 选择收藏夹
const selectCollection = async (collection: any) => {
  collectionStore.setCurrentCollection(collection)
  selectedFolder.value = null
  sidebarCollapsed.value = true
  await collectionStore.fetchFolders(collection.id)
  await collectionStore.fetchBookmarks({ collectionId: collection.id })
}

// 选择文件夹
const selectFolder = async (folderId: string | null) => {
  selectedFolder.value = folderId
  sidebarCollapsed.value = true
  if (collectionStore.currentCollection) {
    await collectionStore.fetchBookmarks({
      collectionId: collectionStore.currentCollection.id,
      folderId: folderId || undefined,
    })
  }
}

// 创建收藏夹
const handleCreateCollection = async () => {
  if (!authStore.isAuthenticated) {
    showToast('请先登录后再进行操作')
    return false
  }
  if (!newCollection.name) {
    showToast('请输入收藏夹名称')
    return false
  }
  const result = await collectionStore.createCollection(newCollection)
  if (result.success) {
    showToast('创建成功')
    newCollection.name = ''
    newCollection.description = ''
    newCollection.icon = ''
  } else {
    showToast(result.message || '创建失败')
  }
  return result.success
}

// 创建文件夹
const handleCreateFolder = async () => {
  if (!authStore.isAuthenticated) {
    showToast('请先登录后再进行操作')
    return false
  }
  if (!newFolder.name || !collectionStore.currentCollection) {
    showToast('请输入文件夹名称')
    return false
  }
  const result = await collectionStore.createFolder({
    name: newFolder.name,
    collectionId: collectionStore.currentCollection.id,
  })
  if (result.success) {
    showToast('创建成功')
    newFolder.name = ''
  } else {
    showToast(result.message || '创建失败')
  }
  return result.success
}

// 重置书签表单
const resetBookmarkForm = () => {
  bookmarkForm.title = ''
  bookmarkForm.url = ''
  bookmarkForm.description = ''
  bookmarkForm.icon = ''
  bookmarkForm.collectionId = undefined
  bookmarkForm.folderId = undefined
  editingBookmark.value = false
}

// 确认添加书签
const handleBookmarkConfirm = async () => {
  try {
    await bookmarkFormRef.value?.validate()
  } catch {
    return
  }
  if (!bookmarkForm.collectionId) {
    showToast('请选择收藏夹')
    return
  }
  const data = {
    title: bookmarkForm.title,
    url: bookmarkForm.url,
    description: bookmarkForm.description,
    icon: bookmarkForm.icon,
    collectionId: bookmarkForm.collectionId,
    folderId: bookmarkForm.folderId || undefined,
  }
  const result = await collectionStore.createBookmark(data)
  if (result.success) {
    showToast('保存成功')
    showCreateBookmark.value = false
    resetBookmarkForm()
  } else {
    showToast(result.message || '保存失败')
  }
}

// 退出登录
const handleLogout = () => {
  showDialog({
    title: '确认退出',
    message: '确定要退出登录吗？',
  }).then(() => {
    authStore.logout()
    router.push('/login')
    showToast('已退出登录')
  })
}

// 统一的 Dialog 方法
const showDialog = (options: any): Promise<void> => {
  if (deviceType.value === 'mobile') {
    return showVantDialog(options).then(() => {})
  } else {
    return ElMessageBox.confirm(options.message, options.title || '提示', {
      confirmButtonText: options.confirmButtonText || '确认',
      cancelButtonText: options.cancelButtonText || '取消',
      type: 'warning',
    }).then(() => {})
  }
}
</script>

<style scoped>
.app-layout {
  position: relative;
  height: 100vh;
  height: 100dvh;
  overflow: hidden;
  background: #f7f8fa;
}

/* 遮罩层 */
.sidebar-overlay {
  position: fixed;
  inset: 0;
  z-index: 100;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(2px);
}

/* 侧边栏 */
.sidebar {
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: 200;
  width: 280px;
  max-width: 85vw;
  background: #fff;
  display: flex;
  flex-direction: column;
  box-shadow: 8px 0 32px rgba(0, 0, 0, 0.12);
}

.sidebar-header {
  padding: 20px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.sidebar-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: -0.3px;
}

.close-btn {
  font-size: 20px;
  color: rgba(255, 255, 255, 0.9);
  cursor: pointer;
  padding: 4px;
  border-radius: 8px;
  transition: background 0.2s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.15);
}

.sidebar-content {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  -webkit-overflow-scrolling: touch;
}

/* 登录提示 */
.login-prompt {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 28px 16px 20px;
  margin-bottom: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  color: #fff;
}

.login-avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
}

.login-prompt p {
  margin: 0 0 16px 0;
  font-size: 14px;
  font-weight: 500;
  opacity: 0.95;
}

.login-btn {
  font-weight: 600;
}

/* 用户信息 */
.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  margin-bottom: 12px;
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  border-radius: 14px;
  color: #fff;
  box-shadow: 0 4px 16px rgba(245, 87, 108, 0.25);
}

.user-avatar {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.25);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.user-details {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-weight: 600;
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-role {
  font-size: 12px;
  opacity: 0.85;
  margin-top: 2px;
}

/* 菜单区域 */
.menu-section {
  margin-bottom: 8px;
}

.section-label {
  font-size: 11px;
  font-weight: 600;
  color: #999;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 8px 12px 6px;
}

.menu-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 11px 12px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 14px;
  color: #333;
}

.menu-item:active {
  transform: scale(0.98);
  background: #f2f3f5;
}

.menu-item.active {
  background: linear-gradient(135deg, #e6f7ff 0%, #bae7ff 100%);
  color: #1677ff;
  font-weight: 600;
}

.menu-icon {
  font-size: 18px;
  width: 24px;
  text-align: center;
  flex-shrink: 0;
}

.folder-custom-icon {
  font-size: 18px;
  width: 24px;
  text-align: center;
  flex-shrink: 0;
}

.menu-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.check-icon {
  font-size: 16px;
  color: #1677ff;
}

.add-item {
  color: #999;
  border: 1px dashed #e0e0e0;
}

.add-item .menu-icon {
  font-size: 16px;
  font-weight: 300;
}

/* 侧边栏底部 */
.sidebar-footer {
  padding: 12px 16px;
  padding-bottom: calc(12px + env(safe-area-inset-bottom, 0px));
  border-top: 1px solid #f2f3f5;
}

/* 主内容区 */
.main-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  height: 100dvh;
}

/* 顶部导航 */
.top-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  padding-top: calc(6px + env(safe-area-inset-top, 0px));
  background: #fff;
  border-bottom: 1px solid #f2f3f5;
  flex-shrink: 0;
}

.menu-btn {
  font-size: 22px;
  color: #333;
  cursor: pointer;
  padding: 6px;
  border-radius: 8px;
  transition: background 0.2s;
  flex-shrink: 0;
}

.menu-btn:active {
  background: #f2f3f5;
}

.top-bar :deep(.van-search) {
  padding: 0;
  flex: 1;
}

.top-bar :deep(.van-search__content) {
  background: #f7f8fa;
  border-radius: 20px;
}

.top-bar :deep(.van-search__content:has(input:focus)) {
  background: #f0f1f3;
}

/* 内容区域 */
.content-area {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  -webkit-overflow-scrolling: touch;
}

/* 浮动按钮样式覆盖 */
/* 滚动条 */
.sidebar-content::-webkit-scrollbar,
.content-area::-webkit-scrollbar {
  width: 3px;
}

.sidebar-content::-webkit-scrollbar-thumb,
.content-area::-webkit-scrollbar-thumb {
  background: #e0e0e0;
  border-radius: 3px;
}

/* 动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-enter-active,
.slide-leave-active {
  transition: transform 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
}

.slide-enter-from,
.slide-leave-to {
  transform: translateX(-100%);
}

/* 桌面端适配 */
@media (min-width: 768px) {
  .sidebar {
    width: 300px;
  }
}

/* 顶栏添加按钮 */
.add-btn {
  font-size: 22px;
  color: #667eea;
  cursor: pointer;
  padding: 6px;
  border-radius: 8px;
  transition: all 0.2s;
  flex-shrink: 0;
}

.add-btn:active {
  transform: scale(0.88);
  background: rgba(102, 126, 234, 0.1);
}

/* 添加书签弹窗样式优化 */
:deep(.bookmark-dialog) {
  .van-dialog__content {
    padding: 12px 0 0;
  }

  .van-field {
    padding: 8px 16px;

    &__control {
      font-size: 15px;
    }

    &__placeholder {
      font-size: 15px;
      color: #c8c9cc;
    }
  }

  .van-field--required::before {
    left: 4px;
    font-size: 12px;
  }

  .van-field textarea {
    font-size: 15px;
  }

  .van-field .van-field__label {
    display: none;
  }
}
</style>
