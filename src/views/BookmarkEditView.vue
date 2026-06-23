<template>
  <div class="bookmark-edit-view">
    <!-- 页面头部 -->
    <div class="edit-header">
      <div class="header-left">
        <el-button :icon="ArrowLeft" @click="goBack">返回</el-button>
        <h2 class="edit-title">
          <el-icon><EditPen /></el-icon>
          {{ isEditing ? '编辑书签' : '添加书签' }}
        </h2>
      </div>
    </div>

    <!-- 表单区域 -->
    <div class="edit-form-wrapper">
      <el-form
        ref="formRef"
        :model="bookmarkForm"
        :rules="formRules"
        label-width="100px"
        label-position="right"
        size="large"
        class="edit-form"
      >
        <!-- 基本信息 -->
        <div class="form-section">
          <div class="section-title">
            <el-icon><InfoFilled /></el-icon>
            基本信息
          </div>

          <el-form-item prop="title">
            <label for="title">标题</label>
            <el-input
              id="title"
              v-model="bookmarkForm.title"
              placeholder="请输入书签标题"
              maxlength="100"
              show-word-limit
            />
          </el-form-item>

          <el-form-item label="URL" prop="url">
            <el-input v-model="bookmarkForm.url" placeholder="请输入网址，如 https://example.com">
              <template #prefix>
                <el-icon><Link /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item label="描述" prop="description">
            <el-input
              v-model="bookmarkForm.description"
              type="textarea"
              placeholder="请输入书签描述（可选）"
              :rows="2"
              maxlength="500"
              show-word-limit
            />
          </el-form-item>

          <el-form-item label="图标">
            <div class="icon-picker">
              <el-input
                v-model="bookmarkForm.icon"
                placeholder="输入 emoji 图标或网址（可选）"
                style="width: 200px"
              />
              <div class="icon-preview" v-if="bookmarkForm.icon">
                <template v-if="isUrl(bookmarkForm.icon)">
                  <img
                    :src="bookmarkForm.icon"
                    :alt="bookmarkForm.title || '书签图标'"
                    class="preview-icon-image"
                    @error="handleIconError"
                  />
                  <span v-if="iconLoadError" class="preview-emoji">🔗</span>
                </template>
                <span v-else class="preview-emoji">{{ bookmarkForm.icon }}</span>
              </div>
              <div class="icon-quick-pick">
                <span
                  v-for="emoji in quickEmojis"
                  :key="emoji"
                  class="emoji-option"
                  @click="bookmarkForm.icon = emoji"
                  >{{ emoji }}</span
                >
              </div>
            </div>
          </el-form-item>
        </div>

        <!-- 分类信息 -->
        <div class="form-section">
          <div class="section-title">
            <el-icon><FolderOpened /></el-icon>
            分类信息
          </div>

          <el-form-item label="收藏夹" prop="collectionId">
            <el-select
              v-model="bookmarkForm.collectionId"
              placeholder="请选择收藏夹"
              style="width: 100%"
              :disabled="isEditing"
              @change="handleCollectionChange"
            >
              <el-option
                v-for="col in collectionStore.collections"
                :key="col.id"
                :label="col.name"
                :value="col.id"
              >
                <span style="display: flex; align-items: center; gap: 8px">
                  <span>{{ col.icon || '📁' }}</span>
                  <span>{{ col.name }}</span>
                </span>
              </el-option>
            </el-select>
          </el-form-item>

          <el-form-item label="文件夹" v-if="bookmarkForm.collectionId">
            <el-select
              v-model="bookmarkForm.folderId"
              placeholder="请选择文件夹（可选）"
              style="width: 100%"
              clearable
            >
              <el-option
                v-for="folder in collectionStore.folders"
                :key="folder.id"
                :label="folder.name"
                :value="folder.id"
              >
                <span style="display: flex; align-items: center; gap: 8px">
                  <el-icon><Folder /></el-icon>
                  <span>{{ folder.name }}</span>
                </span>
              </el-option>
            </el-select>
          </el-form-item>

          <!-- 精选和快照 -->
          <el-form-item>
            <div class="features-row">
              <div class="feature-card">
                <div class="feature-card-body">
                  <div class="feature-card-icon">
                    <el-icon v-if="bookmarkForm.isFeatured"><StarFilled /></el-icon>
                    <el-icon v-else><Star /></el-icon>
                  </div>
                  <div class="feature-card-info">
                    <div class="feature-card-title">精选书签</div>
                    <div class="feature-card-desc">首页突出显示</div>
                  </div>
                  <el-switch
                    v-model="bookmarkForm.isFeatured"
                    style="--el-switch-on-color: #e6a23c"
                  />
                </div>
              </div>
              <div class="feature-card">
                <div class="feature-card-body">
                  <div class="feature-card-icon snapshot-icon">
                    <el-icon><Camera /></el-icon>
                  </div>
                  <div class="feature-card-info">
                    <div class="feature-card-title">页面快照</div>
                    <div class="feature-card-desc">保存网页截图</div>
                    <div
                      v-if="bookmarkForm.hasSnapshot"
                      class="feature-card-actions"
                    >
                      <el-button
                        v-if="bookmarkForm.snapshotUrl"
                        text
                        type="primary"
                        size="small"
                        @click="openSnapshotPreview"
                      >
                        <el-icon><View /></el-icon> 查看
                      </el-button>
                      <el-button
                        text
                        type="default"
                        size="small"
                        @click="handleRegenerateSnapshot"
                        :loading="regeneratingSnapshot"
                      >
                        <el-icon><Refresh /></el-icon> 重新生成
                      </el-button>
                    </div>
                  </div>
                  <el-switch v-model="bookmarkForm.hasSnapshot" />
                </div>
              </div>
            </div>
          </el-form-item>
        </div>

        <!-- 操作按钮 -->
        <div class="form-actions">
          <el-button v-if="isEditing" type="danger" @click="handleDelete" size="large">
            <el-icon><Delete /></el-icon>
            删除书签
          </el-button>
          <div style="flex: 1"></div>
          <el-button @click="goBack" size="large">取消</el-button>
          <el-button type="primary" @click="handleSave" :loading="saving" size="large">
            <el-icon><Check /></el-icon>
            {{ isEditing ? '保存修改' : '添加书签' }}
          </el-button>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import {
  ArrowLeft,
  EditPen,
  InfoFilled,
  Link,
  FolderOpened,
  Folder,
  Check,
  Delete,
  Star,
  StarFilled,
  Camera,
  Refresh,
  View,
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { getFaviconUrl } from '@/utils/ui'
import { useCollectionStore } from '@/stores/collection'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const collectionStore = useCollectionStore()

const formRef = ref<FormInstance>()
const saving = ref(false)
const iconLoadError = ref(false)
const regeneratingSnapshot = ref(false)

// 判断是编辑还是新建
const isEditing = computed(() => {
  const id = route.params.id
  return !!id && id !== 'new'
})

// 快捷 emoji 选择
const quickEmojis = ['🔗', '📌', '⭐', '📝', '💻', '🎨', '📚', '🔧', '🌐', '💡', '🎯', '🚀']

// 表单数据
const bookmarkForm = reactive({
  id: '',
  title: '',
  url: '',
  description: '',
  icon: '',
  collectionId: '',
  folderId: null as string | null,
  isFeatured: false,
  hasSnapshot: false,
  snapshotUrl: '',
})

// 表单校验规则
const formRules = reactive<FormRules>({
  title: [
    { required: true, message: '请输入书签标题', trigger: 'blur' },
    { min: 1, max: 100, message: '标题长度为 1-100 个字符', trigger: 'blur' },
  ],
  url: [
    { required: true, message: '请输入网址', trigger: 'blur' },
    {
      pattern: /^https?:\/\/.+/i,
      message: '请输入有效的网址，以 http:// 或 https:// 开头',
      trigger: 'blur',
    },
  ],
  collectionId: [{ required: true, message: '请选择收藏夹', trigger: 'change' }],
})

// 初始化
onMounted(async () => {
  // 检查登录状态
  if (!authStore.isAuthenticated) {
    ElMessage.warning('请先登录后再进行操作')
    router.push('/login')
    return
  }

  // 加载收藏夹列表
  await collectionStore.fetchCollections()

  // 如果是编辑模式，加载书签数据
  if (route.params.id && route.params.id !== 'new') {
    await loadBookmarkData(route.params.id as string)
  } else {
    // 新建模式：从 query 参数获取预填数据
    if (route.query.collectionId) {
      bookmarkForm.collectionId = route.query.collectionId as string
      await collectionStore.fetchFolders(bookmarkForm.collectionId)
    }
    if (route.query.folderId) {
      bookmarkForm.folderId = route.query.folderId as string
    }
  }
})

// 加载书签数据（编辑模式）
const loadBookmarkData = async (id: string) => {
  // 如果是新建模式，直接返回
  if (id === 'new') return
  try {
    // 重置图标加载错误状态
    resetIconError()
    // 从 store 的 bookmarks 中查找
    const bookmark = collectionStore.bookmarks.find((b) => b.id === id)
    if (bookmark) {
      bookmarkForm.id = bookmark.id
      bookmarkForm.title = bookmark.title
      bookmarkForm.url = bookmark.url
      bookmarkForm.description = bookmark.description || ''
      bookmarkForm.icon = bookmark.icon || ''
      bookmarkForm.collectionId = bookmark.collectionId
      bookmarkForm.folderId = bookmark.folderId || null
      bookmarkForm.isFeatured = bookmark.isFeatured || false
      bookmarkForm.hasSnapshot = bookmark.hasSnapshot || false
      bookmarkForm.snapshotUrl = bookmark.snapshotUrl || ''
      // 加载对应收藏夹的文件夹列表
      await collectionStore.fetchFolders(bookmark.collectionId)
    } else {
      // 如果 store 中没有，尝试通过 API 获取
      const { bookmarkAPI } = await import('@/api')
      const response: any = await bookmarkAPI.getBookmarkById(id)
      if (response?.code === 200 || response?.success) {
        const data = response.data
        bookmarkForm.id = data.id
        bookmarkForm.title = data.title
        bookmarkForm.url = data.url
        bookmarkForm.description = data.description || ''
        bookmarkForm.icon = data.icon || ''
        bookmarkForm.collectionId = data.collectionId
        bookmarkForm.folderId = data.folderId || null
        bookmarkForm.isFeatured = data.isFeatured || false
        bookmarkForm.hasSnapshot = data.hasSnapshot || false
        bookmarkForm.snapshotUrl = data.snapshotUrl || ''
        await collectionStore.fetchFolders(data.collectionId)
      } else {
        ElMessage.error('书签不存在')
        router.push('/')
      }
    }
  } catch (error) {
    console.error('加载书签失败:', error)
    ElMessage.error('加载书签失败')
    router.push('/')
  }
}

// 收藏夹变更处理
const handleCollectionChange = async (collectionId: string) => {
  bookmarkForm.folderId = null
  await collectionStore.fetchFolders(collectionId)
}

// 保存书签
const handleSave = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch {
    return
  }

  if (!authStore.isAuthenticated) {
    ElMessage.warning('请先登录后再进行操作')
    return
  }

  saving.value = true
  try {
    let result
    if (isEditing.value) {
      result = await collectionStore.updateBookmark(bookmarkForm.id, {
        title: bookmarkForm.title,
        url: bookmarkForm.url,
        description: bookmarkForm.description,
        icon: bookmarkForm.icon,
        folderId: bookmarkForm.folderId,
        isFeatured: bookmarkForm.isFeatured,
        hasSnapshot: bookmarkForm.hasSnapshot,
        snapshotUrl: bookmarkForm.snapshotUrl,
      })
    } else {
      result = await collectionStore.createBookmark({
        title: bookmarkForm.title,
        url: bookmarkForm.url,
        description: bookmarkForm.description,
        icon: bookmarkForm.icon,
        collectionId: bookmarkForm.collectionId,
        folderId: bookmarkForm.folderId,
        isFeatured: bookmarkForm.isFeatured,
        hasSnapshot: bookmarkForm.hasSnapshot,
        snapshotUrl: bookmarkForm.snapshotUrl,
      })
    }

    if (result?.success) {
      ElMessage.success(isEditing.value ? '更新成功' : '添加成功')
      goBack()
    } else {
      ElMessage.error(result?.message || '操作失败')
    }
  } finally {
    saving.value = false
  }
}

// 删除书签
const handleDelete = async () => {
  try {
    await ElMessageBox.confirm('确定要删除该书签吗？此操作不可撤销。', '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    const result = await collectionStore.deleteBookmark(bookmarkForm.id)
    if (result.success) {
      ElMessage.success('删除成功')
      goBack()
    } else {
      ElMessage.error(result.message || '删除失败')
    }
  } catch {
    // 用户取消操作
  }
}

// 检查是否是URL
const isUrl = (str: string): boolean => {
  try {
    new URL(str)
    return true
  } catch {
    return false
  }
}

// 重置图标加载错误状态
const resetIconError = () => {
  iconLoadError.value = false
}

// 处理图标加载错误
const handleIconError = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.style.display = 'none'
  // 标记图标加载失败
  iconLoadError.value = true

  // 如果是网址图标加载失败，尝试获取网站favicon
  if (isUrl(bookmarkForm.icon)) {
    try {
      const faviconUrl = getFaviconUrl(bookmarkForm.icon)
      if (faviconUrl && faviconUrl !== bookmarkForm.icon) {
        bookmarkForm.icon = faviconUrl
        iconLoadError.value = false
      }
    } catch (error) {
      console.error('获取favicon失败:', error)
    }
  }
}

// 返回上一页
const goBack = () => {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/')
  }
}

// 打开快照预览
const openSnapshotPreview = () => {
  if (bookmarkForm.snapshotUrl) {
    window.open(bookmarkForm.snapshotUrl, '_blank')
  } else {
    ElMessage.warning('快照文件不存在')
  }
}

// 重新生成快照
const handleRegenerateSnapshot = async () => {
  if (!bookmarkForm.id) {
    ElMessage.warning('请先保存书签后再生成快照')
    return
  }

  regeneratingSnapshot.value = true
  try {
    const { bookmarkAPI } = await import('@/api')
    const response: any = await bookmarkAPI.generateSnapshot(bookmarkForm.id)

    if (response?.code === 200 || response?.success) {
      bookmarkForm.snapshotUrl = response.data.snapshotUrl
      ElMessage.success('快照生成成功')
    } else {
      ElMessage.error(response?.message || '快照生成失败')
    }
  } catch (error: any) {
    console.error('生成快照失败:', error)
    ElMessage.error(error.response?.data?.message || '快照生成失败')
  } finally {
    regeneratingSnapshot.value = false
  }
}
</script>

<style scoped>
.bookmark-edit-view {
  padding: 1px;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow-y: auto;
}

.edit-header {
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

.edit-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #333;
  display: flex;
  align-items: center;
  gap: 6px;
}

.edit-title .el-icon {
  color: #667eea;
}

.edit-form-wrapper {
  max-width: 100%;
  margin: 0;
}

.edit-form {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.form-section {
  background: white;
  border-radius: 8px;
  padding: 12px 16px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
  margin-top: 4px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
  padding-bottom: 4px;
  border-bottom: 1px solid #f0f0f0;
}

.section-title .el-icon {
  color: #667eea;
}

.icon-picker {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.preview-icon-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
  border-radius: 8px;
  position: absolute;
  top: 0;
  left: 0;
}

.icon-preview {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: linear-gradient(135deg, #f0f0f0, #e8e8e8);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
  position: relative;
  overflow: hidden;
}

.preview-emoji {
  font-size: 22px;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: flex;
  align-items: center;
  justify-content: center;
}

.fallback-hidden {
  display: none;
}

.icon-quick-pick {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.emoji-option {
  font-size: 18px;
  cursor: pointer;
  padding: 3px 5px;
  border-radius: 4px;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

.emoji-option:hover {
  background: #f0f0f0;
  border-color: #d9d9d9;
  transform: scale(1.1);
}

.feature-card {
  flex: 1;
  min-width: 0;
  border: 1px solid #e8e8e8;
  border-radius: 10px;
  overflow: hidden;
  transition: border-color 0.2s;
}

.feature-card:hover {
  border-color: #d0d0d0;
}

.feature-card-body {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
}

.features-row {
  display: flex;
  gap: 12px;
  width: 100%;
}

.feature-card-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, #fff7e6, #ffe7ba);
  color: #e6a23c;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}

.feature-card-icon.snapshot-icon {
  background: linear-gradient(135deg, #e6f7ff, #bae7ff);
  color: #409eff;
}

.feature-card-info {
  flex: 1;
  min-width: 0;
}

.feature-card-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  line-height: 1.3;
}

.feature-card-desc {
  font-size: 11px;
  color: #909399;
  line-height: 1.3;
}

.feature-card-actions {
  display: flex;
  gap: 2px;
  margin-top: 2px;
}

/* 操作按钮 */
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
  margin-top: 4px;
}

/* 响应式 */
@media (max-width: 1024px) {
  .edit-form-wrapper {
    max-width: 100%;
  }

  .form-section {
    padding: 16px;
  }
}

@media (max-width: 768px) {
  .bookmark-edit-view {
    padding: 1px;
  }

  .edit-header {
    padding: 12px 16px;
    flex-wrap: wrap;
  }

  .form-section {
    padding: 12px;
  }

  .icon-picker {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
