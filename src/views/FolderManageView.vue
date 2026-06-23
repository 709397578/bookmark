<template>
  <div class="folder-manage-view">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <el-button :icon="ArrowLeft" @click="goBack">返回</el-button>
        <h2 class="page-title">
          <el-icon><Folder /></el-icon>
          文件夹管理
        </h2>
      </div>
      <div class="header-right">
        <el-button type="primary" :icon="Plus" @click="handleCreateFolder">新建文件夹</el-button>
      </div>
    </div>

    <!-- 筛选区域 -->
    <div class="filter-section">
      <el-select
        v-model="selectedCollectionId"
        placeholder="选择收藏夹"
        style="width: 200px"
        @change="handleCollectionChange"
      >
        <el-option
          v-for="collection in collections.filter(
            (c: Collection) => c.isPublic || authStore.isAuthenticated,
          )"
          :key="collection.id"
          :label="collection.name"
          :value="collection.id"
        />
      </el-select>

      <div class="filter-stats">
        共 <strong>{{ folders.length }}</strong> 个文件夹
      </div>
    </div>

    <!-- 文件夹列表 -->
    <div class="folder-table-wrapper" v-loading="loading">
      <el-table
        v-if="folders.length > 0"
        :data="folders"
        stripe
        row-key="id"
        style="width: 100%"
        height="100%"
        empty-text="暂无文件夹"
        :header-cell-style="{ background: '#fafafa', color: '#333', fontWeight: 600 }"
      >
        <el-table-column label="" width="40" align="center" class-name="drag-handle-col">
          <template #default>
            <span class="drag-handle" title="拖拽排序">⠿</span>
          </template>
        </el-table-column>
        <el-table-column label="图标" width="60" align="center">
          <template #default="{ row }">
            <span class="folder-icon">{{ row.icon || '📁' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="颜色" width="80" align="center">
          <template #default="{ row }">
            <div v-if="row.color" class="color-dot" :style="{ backgroundColor: row.color }"></div>
            <span v-else class="no-color">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="folder-name-cell">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="书签数" width="90" align="center">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.bookmarkCount ?? 0 }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170" sortable>
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-button
              size="small"
              type="primary"
              plain
              :icon="Edit"
              @click="handleEditFolder(row)"
            >
              编辑
            </el-button>
            <el-button
              size="small"
              type="danger"
              plain
              :icon="Delete"
              @click="handleDeleteFolder(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-else description="暂无文件夹，点击右上角添加" />
    </div>

    <!-- 创建/编辑文件夹对话框 -->
    <el-dialog
      v-model="showCreateFolder"
      :title="editingFolder ? '编辑文件夹' : '创建文件夹'"
      width="500px"
    >
      <el-form :model="folderForm" label-width="80px">
        <el-form-item label="名称" required for="folder-name">
          <el-input id="folder-name" v-model="folderForm.name" placeholder="请输入文件夹名称" />
        </el-form-item>
        <el-form-item label="图标" for="folder-icon">
          <div class="icon-picker">
            <div class="icon-preview" v-if="folderForm.icon">
              <span class="preview-emoji">{{ folderForm.icon }}</span>
            </div>
            <el-popover trigger="click" :width="340" placement="bottom-start">
              <template #reference>
                <el-button size="small">
                  {{ folderForm.icon ? '更换图标' : '选择图标' }}
                </el-button>
              </template>
              <div class="emoji-picker">
                <div
                  v-for="(emojis, category) in emojiCategories"
                  :key="category"
                  class="emoji-group"
                >
                  <div class="emoji-group-title">{{ category }}</div>
                  <div class="emoji-grid">
                    <span
                      v-for="emoji in emojis"
                      :key="emoji"
                      class="emoji-item"
                      :class="{ active: folderForm.icon === emoji }"
                      @click="folderForm.icon = emoji"
                      >{{ emoji }}</span
                    >
                  </div>
                </div>
              </div>
            </el-popover>
            <el-button
              v-if="folderForm.icon"
              text
              type="danger"
              size="small"
              @click="folderForm.icon = ''"
            >
              清除
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="颜色" for="folder-color">
          <el-color-picker
            id="folder-color"
            v-model="folderForm.color"
            show-alpha
            :predefine="predefineColors"
          />
          <span class="color-hint">选择文件夹颜色（可选）</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateFolder = false">取消</el-button>
        <el-button type="primary" @click="handleSaveFolder">确定</el-button>
      </template>
    </el-dialog>

    <!-- 删除确认对话框 -->
    <el-dialog v-model="showDeleteConfirm" title="确认删除" width="400px">
      <p>确定要删除文件夹 "{{ deletingFolder?.name }}" 吗？此操作不可恢复。</p>
      <template #footer>
        <el-button @click="showDeleteConfirm = false">取消</el-button>
        <el-button type="danger" @click="confirmDeleteFolder">确定删除</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import Sortable from 'sortablejs'
import {
  ArrowLeft,
  Folder,
  Plus,
  Edit,
  Delete,
  Collection as CollectionIcon,
} from '@element-plus/icons-vue'
import { useCollectionStore, type Collection, type Folder as FolderType } from '@/stores/collection'

import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const collectionStore = useCollectionStore()

// 对话框状态
const showCreateFolder = ref(false)
const showDeleteConfirm = ref(false)
const selectedCollectionId = ref('')

// 预定义颜色列表
const predefineColors = ref([
  '#ff4500',
  '#ff8c00',
  '#ffd700',
  '#90ee90',
  '#00ced1',
  '#1e90ff',
  '#c71585',
  '#ff69b4',
  '#ff6347',
  '#7b68ee',
  '#00fa9a',
  '#ffdab9',
  '#87ceeb',
  '#da70d6',
  '#32cd32',
  '#fa8072',
])

// 文件夹表单
const folderForm = ref({
  id: '',
  name: '',
  icon: '',
  color: '',
  collectionId: '',
})

// 文件夹图标分类
const emojiCategories: Record<string, string[]> = {
  常用: ['📁', '📂', '⭐', '📌', '💎', '🔥', '💡', '🎯', '🚀', '💎', '🏆', '🎉'],
  工作: ['💼', '📊', '📈', '📋', '📝', '📎', '🔍', '🔧', '⚙️', '💻', '🖥️', '🖨️'],
  学习: ['📚', '📖', '🎓', '🔬', '🧪', '📐', '📏', '✏️', '🖊️', '🧠', '💡', '📖'],
  媒体: ['🎬', '🎵', '📷', '🎨', '🖼️', '📺', '🎸', '🎤', '🎧', '📸', '🎭', '🎪'],
  社交: ['💬', '👥', '🌐', '🔗', '📧', '💌', '📱', '📞', '🤝', '👥', '💬', '📣'],
  生活: ['🏠', '🛒', '✈️', '🚗', '🍕', '☕', '🎮', '⚽', '🎵', '🌸', '🌈', '☀️'],
  技术: ['💻', '🖥️', '⌨️', '🖱️', '📡', '🔌', '💾', '📀', '🐛', '🛠️', '⚙️', '🔧'],
}

// 当前操作的文件夹
const editingFolder = ref<any>(null)
const deletingFolder = ref<any>(null)

// 获取收藏夹列表
const collections = ref<Collection[]>([])

// 获取文件夹列表
const folders = ref<FolderType[]>([])
const loading = ref(false)

// 初始化
onMounted(async () => {
  if (!authStore.isAuthenticated) {
    ElMessage.warning('请先登录后再进行操作')
    router.push('/login')
    return
  }

  await loadCollections()
})

// 加载收藏夹列表
const loadCollections = async () => {
  try {
    // 根据登录状态决定是否只获取公开收藏夹
    const publicOnly = !authStore.isAuthenticated
    await collectionStore.fetchCollections(publicOnly)
    collections.value = collectionStore.collections
    // 默认选择第一个收藏夹
    if (collections.value.length > 0 && !selectedCollectionId.value) {
      selectedCollectionId.value = collections.value[0]!.id
      await loadFolders()
    }
  } catch (error) {
    console.error('加载收藏夹失败:', error)
    ElMessage.error('加载收藏夹失败')
  }
}

// 收藏夹切换处理
const handleCollectionChange = async (collectionId: string) => {
  selectedCollectionId.value = collectionId
  await loadFolders()
}

// 加载文件夹列表
const loadFolders = async () => {
  if (!selectedCollectionId.value) return

  loading.value = true
  try {
    await collectionStore.fetchFolders(selectedCollectionId.value)
    folders.value = collectionStore.folders
  } catch (error) {
    console.error('加载文件夹失败:', error)
    ElMessage.error('加载文件夹失败')
  } finally {
    loading.value = false
  }
}

// 打开创建文件夹对话框
const handleCreateFolder = () => {
  folderForm.value = {
    id: '',
    name: '',
    icon: '',
    color: '',
    collectionId: selectedCollectionId.value,
  }
  editingFolder.value = null
  showCreateFolder.value = true
}

// 打开编辑文件夹对话框
const handleEditFolder = (folder: any) => {
  folderForm.value = { ...folder }
  // 确保color字段有默认值
  if (!folderForm.value.color) {
    folderForm.value.color = ''
  }
  editingFolder.value = folder
  showCreateFolder.value = true
}

// 保存文件夹
const handleSaveFolder = async () => {
  if (!folderForm.value.name) {
    ElMessage.warning('请输入文件夹名称')
    return
  }

  try {
    if (editingFolder.value) {
      // 更新文件夹
      const updateData: any = {
        name: folderForm.value.name,
        icon: folderForm.value.icon,
      }

      // 只在有颜色值时添加color字段
      if (
        folderForm.value.color !== undefined &&
        folderForm.value.color !== null &&
        folderForm.value.color !== ''
      ) {
        updateData.color = folderForm.value.color
      }

      const result = await collectionStore.updateFolder(folderForm.value.id, updateData)
      if (result.success) {
        ElMessage.success('更新成功')
        await loadFolders()
      } else {
        ElMessage.error(result.message || '更新失败')
      }
    } else {
      // 创建文件夹
      const createData: any = {
        name: folderForm.value.name,
        icon: folderForm.value.icon,
        collectionId: selectedCollectionId.value,
      }

      // 只在有颜色值时添加color字段
      if (
        folderForm.value.color !== undefined &&
        folderForm.value.color !== null &&
        folderForm.value.color !== ''
      ) {
        createData.color = folderForm.value.color
      }

      const result = await collectionStore.createFolder(createData)
      if (result.success) {
        ElMessage.success('创建成功')
        await loadFolders()
      } else {
        ElMessage.error(result.message || '创建失败')
      }
    }
    // 重置表单
    folderForm.value = {
      id: '',
      name: '',
      icon: '',
      color: '',
      collectionId: selectedCollectionId.value,
    }
    editingFolder.value = null
    showCreateFolder.value = false
  } catch (error) {
    console.error('保存文件夹失败:', error)
    ElMessage.error('保存失败')
  }
}

// 删除文件夹
const handleDeleteFolder = (folder: any) => {
  deletingFolder.value = folder
  showDeleteConfirm.value = true
}

// 确认删除文件夹
const confirmDeleteFolder = async () => {
  if (!deletingFolder.value) return

  try {
    const result = await collectionStore.deleteFolder(deletingFolder.value.id)
    if (result.success) {
      ElMessage.success('删除成功')
      await loadFolders()
    } else {
      ElMessage.error(result.message || '删除失败')
    }
    showDeleteConfirm.value = false
    deletingFolder.value = null
  } catch (error) {
    console.error('删除文件夹失败:', error)
    ElMessage.error('删除失败')
    showDeleteConfirm.value = false
    deletingFolder.value = null
  }
}

// 初始化
onMounted(async () => {
  await loadCollections()
})

// 格式化日期
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

// 返回上一层
const goBack = () => {
  router.back()
}

// 拖拽排序
let sortableInstance: Sortable | null = null
const initSortable = async () => {
  await nextTick()
  const tableBody = document.querySelector(
    '.folder-table-wrapper .el-table__body-wrapper tbody',
  ) as HTMLElement | null
  if (!tableBody) return
  if (sortableInstance) sortableInstance.destroy()
  sortableInstance = Sortable.create(tableBody, {
    handle: '.drag-handle',
    animation: 200,
    ghostClass: 'sortable-ghost',
    onEnd: async (evt: Sortable.SortableEvent) => {
      const { oldIndex, newIndex } = evt
      if (oldIndex === undefined || newIndex === undefined || oldIndex === newIndex) return
      const list = [...folders.value]
      const moved = list.splice(oldIndex, 1)[0]
      if (!moved) return
      list.splice(newIndex, 0, moved)
      const orders = list.map((item, index) => ({ id: item.id, order: index }))
      folders.value = list.map((item, index) => ({
        ...item,
        sortOrder: index,
      }))
      await collectionStore.batchUpdateFolderSortOrders(orders)
    },
  })
}

watch(
  () => folders.value.length,
  () => {
    initSortable()
  },
)

onBeforeUnmount(() => {
  if (sortableInstance) sortableInstance.destroy()
})
</script>

<style scoped>
.folder-manage-view {
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

.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
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

.folder-table-wrapper {
  flex: 1;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
  overflow: auto;
  padding: 12px;
  margin-top: 4px;
  max-height: calc(100vh - 200px);
}

.folder-name-cell {
  font-weight: 500;
  color: #333;
}

.folder-icon {
  font-size: 18px;
  display: inline-block;
}

.color-dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  margin: 0 auto;
  box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.1);
}

.no-color {
  color: #999;
  font-size: 12px;
}

.drag-handle {
  cursor: grab;
  color: #c0c4cc;
  font-size: 16px;
  user-select: none;
  transition: color 0.2s;
}

.drag-handle:hover {
  color: #667eea;
  cursor: grabbing;
}

:deep(.sortable-ghost) {
  opacity: 0.4;
  background: #e6f7ff !important;
}

.color-hint {
  margin-left: 8px;
  font-size: 12px;
  color: #666;
}

.icon-picker {
  display: flex;
  align-items: center;
  gap: 10px;
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
  flex-shrink: 0;
}

.preview-emoji {
  font-size: 22px;
}

.emoji-picker {
  max-height: 320px;
  overflow-y: auto;
}

.emoji-group {
  margin-bottom: 8px;
}

.emoji-group-title {
  font-size: 12px;
  color: #909399;
  font-weight: 600;
  margin-bottom: 6px;
  padding-left: 2px;
}

.emoji-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 4px;
}

.emoji-item {
  font-size: 20px;
  cursor: pointer;
  padding: 4px;
  border-radius: 6px;
  text-align: center;
  transition: all 0.15s;
  border: 2px solid transparent;
}

.emoji-item:hover {
  background: #f0f0f0;
  transform: scale(1.15);
}

.emoji-item.active {
  background: #ecf5ff;
  border-color: #409eff;
}

/* 响应式 */
@media (max-width: 1024px) {
  .page-header {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }

  .header-right {
    flex-wrap: wrap;
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
