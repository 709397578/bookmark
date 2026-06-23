<template>
  <div class="bookmark-manage-view">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <el-button :icon="ArrowLeft" @click="$router.push('/')">返回首页</el-button>
        <h2 class="page-title">
          <el-icon><Collection /></el-icon>
          书签管理
        </h2>
      </div>
      <div class="header-right">
        <el-button
          type="primary"
          :icon="Plus"
          @click="
            router.push({
              name: 'bookmark-edit',
              params: { id: 'new' },
              query: { collectionId: managerCollectionId },
            })
          "
          >添加书签</el-button
        >
        <el-button :icon="FolderAdd" @click="showCreateCollection = true">新建收藏夹</el-button>
        <el-button :icon="Collection" @click="showCollectionManagement = true"
          >收藏夹管理</el-button
        >
        <el-button :icon="Folder" @click="showCreateFolder = true">新建文件夹</el-button>
        <el-button :icon="Files" @click="router.push({ name: 'folder-manage' })"
          >文件夹管理</el-button
        >
        <el-button :icon="Upload" @click="router.push({ name: 'bookmark-import' })"
          >导入书签</el-button
        >
        <el-dropdown :disabled="!managerCollectionId" @command="handleExportCommand">
          <el-button :icon="Download" :disabled="!managerCollectionId">
            导出格式 <el-icon class="el-icon--right"><arrow-down /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="json">JSON 格式</el-dropdown-item>
              <el-dropdown-item command="html">Chrome HTML 格式</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <!-- 筛选区域 -->
    <div class="filter-section">
      <el-input
        v-model="managerSearchQuery"
        placeholder="搜索书签..."
        clearable
        style="width: 300px"
        @input="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>

      <el-select
        v-model="managerCollectionId"
        placeholder="选择收藏夹"
        style="width: 200px"
        @change="handleManagerCollectionChange"
      >
        <el-option
          v-for="col in collectionStore.collections.filter(
            (c) => c.isPublic || authStore.isAuthenticated,
          )"
          :key="col.id"
          :label="col.name"
          :value="col.id"
        />
      </el-select>

      <el-select
        v-model="managerFolderId"
        placeholder="选择文件夹"
        style="width: 200px"
        clearable
        @change="handleManagerFolderChange"
      >
        <el-option
          v-for="folder in collectionStore.folders"
          :key="folder.id"
          :label="folder.name"
          :value="folder.id"
        />
        <el-option label="未分类" value="none" />
      </el-select>

      <div class="filter-stats">
        共 <strong>{{ filteredManagerBookmarks.length }}</strong> 个书签
      </div>
    </div>

    <!-- 书签列表 -->
    <div class="bookmark-table-wrapper" v-loading="collectionStore.loading">
      <!-- 批量操作栏 -->
      <div v-if="selectedBookmarkIds.length > 0" class="batch-bar">
        <span class="batch-count"
          >已选 <strong>{{ selectedBookmarkIds.length }}</strong> 个书签</span
        >
        <el-button size="small" @click="clearSelection">取消全选</el-button>
        <el-button size="small" type="warning" :icon="FolderOpened" @click="showBatchMove = true">
          批量移动
        </el-button>
        <el-button size="small" type="danger" :icon="Delete" @click="confirmBatchDelete">
          批量删除
        </el-button>
      </div>

      <el-table
        v-if="filteredManagerBookmarks.length > 0"
        :data="filteredManagerBookmarks"
        stripe
        style="width: 100%"
        height="100%"
        row-key="id"
        empty-text="暂无书签"
        :header-cell-style="{ background: '#fafafa', color: '#333', fontWeight: 600 }"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column label="" width="40" align="center" class-name="drag-handle-col">
          <template #default>
            <span class="drag-handle" title="拖拽排序">⠿</span>
          </template>
        </el-table-column>
        <el-table-column label="图标" width="60" align="center">
          <template #default="{ row }">
            <img
              v-if="getFaviconUrl(row.url)"
              :src="getFaviconUrl(row.url)"
              :alt="row.title"
              class="bookmark-favicon"
              @load="handleFaviconLoad"
              @error="handleFaviconError"
            />
            <span
              :class="['bookmark-icon-fallback', { 'fallback-hidden': getFaviconUrl(row.url) }]"
              >{{ row.icon || '🔗' }}</span
            >
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="bookmark-title-cell">{{ row.title }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="url" label="URL" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            <el-link :href="row.url" target="_blank" type="primary" underline="never">
              {{ row.url }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip />
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
              @click="openEditBookmark(row)"
            >
              编辑
            </el-button>
            <el-button
              size="small"
              type="danger"
              plain
              :icon="Delete"
              @click="confirmDeleteBookmark(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-else description="暂无书签，点击右上角添加" />
    </div>

    <!-- 创建/编辑收藏夹对话框 -->
    <el-dialog
      v-model="showCreateCollection"
      :title="editingCollection ? '编辑收藏夹' : '创建收藏夹'"
      width="500px"
    >
      <el-form :model="newCollection" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="newCollection.name" placeholder="请输入收藏夹名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="newCollection.description"
            type="textarea"
            placeholder="请输入描述（可选）"
          />
        </el-form-item>
        <el-form-item label="图标">
          <div class="icon-picker">
            <div class="icon-preview" v-if="newCollection.icon">
              <span class="preview-emoji">{{ newCollection.icon }}</span>
            </div>
            <el-popover trigger="click" :width="340" placement="bottom-start">
              <template #reference>
                <el-button size="small">
                  {{ newCollection.icon ? '更换图标' : '选择图标' }}
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
                      :class="{ active: newCollection.icon === emoji }"
                      @click="newCollection.icon = emoji"
                      >{{ emoji }}</span
                    >
                  </div>
                </div>
              </div>
            </el-popover>
            <el-button
              v-if="newCollection.icon"
              text
              type="danger"
              size="small"
              @click="newCollection.icon = ''"
            >
              清除
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="可见性">
          <div style="display: flex; align-items: center; gap: 12px">
            <el-switch
              v-model="newCollection.isPublic"
              active-text="所有人可见"
              inactive-text="仅自己可见"
              inline-prompt
            />
            <el-icon v-if="newCollection.isPublic" color="#67c23a"><InfoFilled /></el-icon>
            <el-icon v-else color="#909399"><InfoFilled /></el-icon>
          </div>
          <div style="font-size: 12px; color: #999; margin-top: 4px">
            {{ newCollection.isPublic ? '其他用户可以查看此收藏夹' : '只有你自己可以查看此收藏夹' }}
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateCollection = false">取消</el-button>
        <el-button type="primary" @click="handleSaveCollection">确定</el-button>
      </template>
    </el-dialog>

    <!-- 收藏夹管理对话框 -->
    <el-dialog v-model="showCollectionManagement" title="收藏夹管理" width="800px">
      <div class="collection-management">
        <el-table :data="collectionStore.collections" stripe row-key="id">
          <el-table-column label="" width="40" align="center" class-name="drag-handle-col">
            <template #default>
              <span class="drag-handle" title="拖拽排序">⠿</span>
            </template>
          </el-table-column>
          <el-table-column label="图标" width="60" align="center">
            <template #default="{ row }">
              <span class="collection-icon">{{ row.icon || '📁' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="名称" prop="name">
            <template #default="{ row }">
              <span class="collection-name">{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column label="描述" prop="description">
            <template #default="{ row }">
              <span class="collection-description">{{ row.description || '无描述' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="书签数" width="90" align="center">
            <template #default="{ row }">
              <el-tag type="info" size="small">{{ row.bookmarkCount ?? 0 }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="公开状态" width="90" align="center">
            <template #default="{ row }">
              <el-tag :type="row.isPublic ? 'success' : 'info'" size="small">
                {{ row.isPublic ? '公开' : '私有' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180" align="center">
            <template #default="{ row }">
              <el-button size="small" type="primary" :icon="Edit" @click="handleEditCollection(row)"
                >编辑</el-button
              >
              <el-button
                size="small"
                type="danger"
                :icon="Delete"
                @click="handleDeleteCollection(row)"
                >删除</el-button
              >
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-dialog>

    <!-- 创建文件夹对话框 -->
    <el-dialog v-model="showCreateFolder" title="创建文件夹" width="500px">
      <el-form :model="newFolder" label-width="80px">
        <el-form-item label="收藏夹" required>
          <el-select
            v-model="newFolder.collectionId"
            placeholder="请选择收藏夹"
            style="width: 100%"
            @change="handleCreateFolderCollectionChange"
          >
            <el-option
              v-for="collection in collectionStore.collections"
              :key="collection.id"
              :label="collection.name"
              :value="collection.id"
            >
              <span style="display: flex; align-items: center; gap: 8px">
                <span>{{ collection.icon || '📁' }}</span>
                <span>{{ collection.name }}</span>
              </span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="newFolder.name" placeholder="请输入文件夹名称" />
        </el-form-item>
        <el-form-item label="图标">
          <div class="icon-picker">
            <div class="icon-preview" v-if="newFolder.icon">
              <span class="preview-emoji">{{ newFolder.icon }}</span>
            </div>
            <el-popover trigger="click" :width="340" placement="bottom-start">
              <template #reference>
                <el-button size="small">
                  {{ newFolder.icon ? '更换图标' : '选择图标' }}
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
                      :class="{ active: newFolder.icon === emoji }"
                      @click="newFolder.icon = emoji"
                      >{{ emoji }}</span
                    >
                  </div>
                </div>
              </div>
            </el-popover>
            <el-button
              v-if="newFolder.icon"
              text
              type="danger"
              size="small"
              @click="newFolder.icon = ''"
            >
              清除
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateFolder = false">取消</el-button>
        <el-button type="primary" @click="handleCreateFolder">确定</el-button>
      </template>
    </el-dialog>

    <!-- 添加/编辑书签对话框 -->
    <router-view :key="$route.fullPath"></router-view>

    <!-- 批量移动对话框 -->
    <el-dialog v-model="showBatchMove" title="批量移动书签" width="400px">
      <el-form label-width="100px">
        <el-form-item label="目标文件夹">
          <el-select
            v-model="batchMoveFolderId"
            placeholder="选择文件夹"
            style="width: 100%"
            clearable
          >
            <el-option
              v-for="folder in collectionStore.folders"
              :key="folder.id"
              :label="folder.name"
              :value="folder.id"
            />
            <el-option label="未分类" value="" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBatchMove = false">取消</el-button>
        <el-button type="primary" @click="handleBatchMove">确定移动</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import Sortable from 'sortablejs'
import {
  ArrowLeft,
  Plus,
  FolderAdd,
  Folder,
  Files,
  Search,
  Collection,
  Edit,
  Delete,
  InfoFilled,
  Link,
  Upload,
  Download,
  FolderOpened,
  Check,
  ArrowDown,
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useCollectionStore } from '@/stores/collection'
import { bookmarkAPI } from '@/api'
import { getFaviconUrl } from '@/utils/ui'
const router = useRouter()
const authStore = useAuthStore()
const collectionStore = useCollectionStore()
// saving变量已移除，保留以避免错误

// 对话框状态
const showCreateCollection = ref(false)
const showCreateFolder = ref(false)
const showCollectionManagement = ref(false)
const editingCollection = ref<any>(null)

// 书签管理相关状态
const managerSearchQuery = ref('')
const managerCollectionId = ref<string>('')
const managerFolderId = ref<string | null>(null)

// 批量操作状态
const selectedBookmarkIds = ref<string[]>([])
const showBatchMove = ref(false)
const batchMoveFolderId = ref('')

// 处理favicon加载成功
const handleFaviconLoad = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.style.display = 'block'
  const fallback = img.nextElementSibling as HTMLElement
  if (fallback) {
    fallback.classList.add('fallback-hidden')
  }
}

// 处理favicon加载错误
const handleFaviconError = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.style.display = 'none'
  const fallback = img.nextElementSibling as HTMLElement
  if (fallback) {
    fallback.classList.remove('fallback-hidden')
  }
}

// 表单数据
const newCollection = reactive({
  name: '',
  description: '',
  icon: '',
  isPublic: false,
})

const newFolder = reactive({
  name: '',
  collectionId: '',
  icon: '',
})

const bookmarkForm = reactive({
  id: '',
  title: '',
  url: '',
  description: '',
  icon: '',
  collectionId: '',
  folderId: null as string | null,
})

const iconLoadError = ref(false)

// 收藏夹图标分类
const emojiCategories: Record<string, string[]> = {
  常用: ['📁', '📂', '⭐', '📌', '💎', '🔥', '💡', '🎯', '🚀', '💎', '🏆', '🎉'],
  工作: ['💼', '📊', '📈', '📋', '📝', '📎', '🔍', '🔧', '⚙️', '💻', '🖥️', '🖨️'],
  学习: ['📚', '📖', '🎓', '🔬', '🧪', '📐', '📏', '✏️', '🖊️', '🧠', '💡', '📖'],
  媒体: ['🎬', '🎵', '📷', '🎨', '🖼️', '📺', '🎸', '🎤', '🎧', '📸', '🎭', '🎪'],
  社交: ['💬', '👥', '🌐', '🔗', '📧', '💌', '📱', '📞', '🤝', '👥', '💬', '📣'],
  生活: ['🏠', '🛒', '✈️', '🚗', '🍕', '☕', '🎮', '⚽', '🎵', '🌸', '🌈', '☀️'],
  技术: ['💻', '🖥️', '⌨️', '🖱️', '📡', '🔌', '💾', '📀', '🐛', '🛠️', '⚙️', '🔧'],
}

// 表单校验规则
const bookmarkFormRules = reactive<FormRules>({
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
  await collectionStore.fetchCollections()
  // 如果有收藏夹，自动选择第一个
  if (collectionStore.collections.length > 0) {
    const firstCol = collectionStore.collections[0]
    if (firstCol) {
      managerCollectionId.value = firstCol.id
      await collectionStore.fetchFolders(firstCol.id)
      await collectionStore.fetchBookmarks({ collectionId: firstCol.id })
    }
  }
})

// 检查是否是URL
const isUrl = (str: string): boolean => {
  try {
    new URL(str)
    return true
  } catch {
    return false
  }
}

// 处理图标加载错误
const handleIconError = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.style.display = 'none'
  // 标记图标加载失败
  iconLoadError.value = true
}

// 重置图标加载错误状态
const resetIconError = () => {
  iconLoadError.value = false
}

// 书签管理：收藏夹切换处理
const handleManagerCollectionChange = async (collectionId: string) => {
  managerFolderId.value = null
  await collectionStore.fetchFolders(collectionId)
  await collectionStore.fetchBookmarks({ collectionId })
}

// 书签管理：文件夹切换处理
const handleManagerFolderChange = async (folderId: string | null) => {
  if (managerCollectionId.value) {
    await collectionStore.fetchBookmarks({
      collectionId: managerCollectionId.value,
      folderId: folderId === 'none' ? undefined : folderId || undefined,
    })
  }
}

// 书签管理：过滤后的书签列表
const filteredManagerBookmarks = computed(() => {
  let bookmarks = collectionStore.bookmarks

  if (managerSearchQuery.value) {
    const query = managerSearchQuery.value.toLowerCase()
    bookmarks = bookmarks.filter(
      (b: any) =>
        b.title?.toLowerCase().includes(query) ||
        b.url?.toLowerCase().includes(query) ||
        b.description?.toLowerCase().includes(query),
    )
  }

  if (managerFolderId.value === 'none') {
    bookmarks = bookmarks.filter((b: any) => !b.folderId)
  } else if (managerFolderId.value) {
    bookmarks = bookmarks.filter((b: any) => b.folderId === managerFolderId.value)
  }

  return bookmarks
})

// 拖拽排序
let sortableInstance: Sortable | null = null
const initSortable = async () => {
  await nextTick()
  const tableBody = document.querySelector(
    '.bookmark-table-wrapper .el-table__body-wrapper tbody',
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
      const list = [...filteredManagerBookmarks.value]
      const moved = list.splice(oldIndex, 1)[0]
      if (!moved) return
      list.splice(newIndex, 0, moved)
      const orders = list.map((item, index) => ({ id: item.id, order: index }))
      collectionStore.bookmarks = list.map((item, index) => ({
        ...item,
        sortOrder: index,
      }))
      await collectionStore.batchUpdateSortOrders(orders)
    },
  })
}

watch(
  () => filteredManagerBookmarks.value.length,
  () => {
    initSortable()
  },
)

onBeforeUnmount(() => {
  if (sortableInstance) sortableInstance.destroy()
  if (collectionSortableInstance) collectionSortableInstance.destroy()
})

// 收藏夹管理拖拽排序
let collectionSortableInstance: Sortable | null = null
const initCollectionSortable = async () => {
  await nextTick()
  const tableBody = document.querySelector(
    '.collection-management .el-table__body-wrapper tbody',
  ) as HTMLElement | null
  if (!tableBody) return
  if (collectionSortableInstance) collectionSortableInstance.destroy()
  collectionSortableInstance = Sortable.create(tableBody, {
    handle: '.drag-handle',
    animation: 200,
    ghostClass: 'sortable-ghost',
    onEnd: async (evt: Sortable.SortableEvent) => {
      const { oldIndex, newIndex } = evt
      if (oldIndex === undefined || newIndex === undefined || oldIndex === newIndex) return
      const list = [...collectionStore.collections]
      const moved = list.splice(oldIndex, 1)[0]
      if (!moved) return
      list.splice(newIndex, 0, moved)
      const orders = list.map((item, index) => ({ id: item.id, order: index }))
      collectionStore.collections = list.map((item, index) => ({
        ...item,
        sortOrder: index,
      }))
      await collectionStore.batchUpdateCollectionSortOrders(orders)
    },
  })
}

watch(showCollectionManagement, (val) => {
  if (val) initCollectionSortable()
})

// 搜索处理
const handleSearch = () => {
  // 搜索由 computed 自动处理
}

// 创建收藏夹
const handleCreateCollection = async () => {
  if (!authStore.isAuthenticated) {
    ElMessage.warning('请先登录后再进行操作')
    return
  }

  if (!newCollection.name) {
    ElMessage.warning('请输入收藏夹名称')
    return
  }
  const result = await collectionStore.createCollection(newCollection)
  if (result.success) {
    ElMessage.success('创建成功')
    resetCollectionForm()
    showCreateCollection.value = false
  } else {
    ElMessage.error(result.message || '创建失败')
  }
}

// 编辑收藏夹
const handleEditCollection = (collection: any) => {
  editingCollection.value = collection
  newCollection.name = collection.name
  newCollection.description = collection.description || ''
  newCollection.icon = collection.icon || ''
  newCollection.isPublic = collection.isPublic
  showCreateCollection.value = true
}

// 保存收藏夹（创建或更新）
const handleSaveCollection = async () => {
  if (!authStore.isAuthenticated) {
    ElMessage.warning('请先登录后再进行操作')
    return
  }

  if (!newCollection.name) {
    ElMessage.warning('请输入收藏夹名称')
    return
  }

  try {
    if (editingCollection.value) {
      // 更新收藏夹
      const result = await collectionStore.updateCollection(editingCollection.value.id, {
        name: newCollection.name,
        description: newCollection.description,
        icon: newCollection.icon,
        isPublic: newCollection.isPublic,
      })
      if (result.success) {
        ElMessage.success('更新成功')
        resetCollectionForm()
        showCreateCollection.value = false
      } else {
        ElMessage.error(result.message || '更新失败')
      }
    } else {
      // 创建收藏夹
      const result = await collectionStore.createCollection(newCollection)
      if (result.success) {
        ElMessage.success('创建成功')
        resetCollectionForm()
        showCreateCollection.value = false
      } else {
        ElMessage.error(result.message || '创建失败')
      }
    }
  } catch (error) {
    console.error('保存收藏夹失败:', error)
    ElMessage.error('保存失败')
  }
}

// 重置收藏夹表单
const resetCollectionForm = () => {
  newCollection.name = ''
  newCollection.description = ''
  newCollection.icon = ''
  newCollection.isPublic = false
  editingCollection.value = null
}

// 删除收藏夹
const handleDeleteCollection = async (collection: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除收藏夹"${collection.name}"吗？此操作不可恢复，且将删除该收藏夹下的所有书签和文件夹。`,
      '确认删除',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )

    const result = await collectionStore.deleteCollection(collection.id)
    if (result.success) {
      ElMessage.success('删除成功')
      // 如果删除的是当前选中的收藏夹，重置选中状态
      if (collectionStore.currentCollection?.id === collection.id) {
        collectionStore.setCurrentCollection(null)
      }
      showCollectionManagement.value = false
    } else {
      ElMessage.error(result.message || '删除失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除收藏夹失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 创建文件夹：收藏夹选择变化处理
const handleCreateFolderCollectionChange = (collectionId: string) => {
  newFolder.collectionId = collectionId
}

// 处理书签表单中收藏夹的选择函数已移除，保留以避免错误

// 创建文件夹
const handleCreateFolder = async () => {
  if (!authStore.isAuthenticated) {
    ElMessage.warning('请先登录后再进行操作')
    return
  }

  if (!newFolder.name || !newFolder.collectionId) {
    ElMessage.warning('请选择收藏夹并输入文件夹名称')
    return
  }

  const result = await collectionStore.createFolder({
    name: newFolder.name,
    collectionId: newFolder.collectionId,
    icon: newFolder.icon || undefined,
  })
  if (result.success) {
    ElMessage.success('创建成功')
    newFolder.name = ''
    newFolder.collectionId = ''
    newFolder.icon = ''
    showCreateFolder.value = false

    // 如果创建的文件夹属于当前选中的收藏夹，刷新文件夹列表
    if (managerCollectionId.value === newFolder.collectionId) {
      await collectionStore.fetchFolders(newFolder.collectionId)
    }
  } else {
    ElMessage.error(result.message || '创建失败')
  }
}

// 保存书签函数已移除，因为现在使用独立的编辑页面
// 保留此函数以避免错误，但实际功能已迁移到BookmarkEditView.vue

// 重置书签表函数已移除，保留以避免错误

// 打开编辑书签对话框
const openEditBookmark = (bookmark: any) => {
  if (!authStore.isAuthenticated) {
    ElMessage.warning('请先登录后再进行操作')
    return
  }

  // 跳转到编辑页面
  router.push({
    name: 'bookmark-edit',
    params: { id: bookmark.id },
    query: {
      collectionId: bookmark.collectionId,
      folderId: bookmark.folderId || undefined,
    },
  })
}

// 打开添加书签对话框
const openAddBookmark = () => {
  if (!authStore.isAuthenticated) {
    ElMessage.warning('请先登录后再进行操作')
    return
  }

  // 跳转到添加页面
  router.push({
    name: 'bookmark-edit',
    params: { id: 'new' },
    query: {
      collectionId: managerCollectionId.value,
    },
  })
}

// 确认删除书签
const confirmDeleteBookmark = async (bookmark: any) => {
  try {
    await ElMessageBox.confirm('确定要删除该书签吗？此操作不可撤销。', '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    const result = await collectionStore.deleteBookmark(bookmark.id)
    if (result.success) {
      ElMessage.success('删除成功')
      if (managerCollectionId.value) {
        await collectionStore.fetchBookmarks({
          collectionId: managerCollectionId.value,
          folderId:
            managerFolderId.value === 'none' ? undefined : managerFolderId.value || undefined,
        })
      }
    } else {
      ElMessage.error(result.message || '删除失败')
    }
  } catch {
    // 用户取消操作，不做处理
  }
}

// 初始化
onMounted(async () => {
  // 根据登录状态决定是否只获取公开收藏夹
  const publicOnly = !authStore.isAuthenticated
  await collectionStore.fetchCollections(publicOnly)

  // 如果有收藏夹，默认选中第一个
  if (collectionStore.collections.length > 0 && !managerCollectionId.value) {
    managerCollectionId.value = collectionStore.collections[0]!.id
    await collectionStore.fetchFolders(managerCollectionId.value)
    await collectionStore.fetchBookmarks({ collectionId: managerCollectionId.value })
  }
})

// 格式化日期
const formatDate = (timeStr: string) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// 导出书签
const handleExportBookmarks = async () => {
  if (!managerCollectionId.value) {
    ElMessage.warning('请先选择收藏夹')
    return
  }

  try {
    const response = await bookmarkAPI.exportBookmarks(managerCollectionId.value)
    const data = response.data

    // 创建JSON文件并下载
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `bookmarks-${new Date().toISOString().slice(0, 10)}.json`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)

    ElMessage.success('导出成功')
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败，请重试')
  }
}

// 导出格式选择
const handleExportCommand = async (command: string) => {
  if (!managerCollectionId.value) {
    ElMessage.warning('请先选择收藏夹')
    return
  }

  if (command === 'html') {
    try {
      const response = await bookmarkAPI.exportBookmarksHTML(managerCollectionId.value)
      const blob = new Blob([response as any], { type: 'text/html;charset=utf-8' })
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `bookmarks-${new Date().toISOString().slice(0, 10)}.html`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
      ElMessage.success('导出成功')
    } catch (error) {
      console.error('导出HTML失败:', error)
      ElMessage.error('导出失败，请重试')
    }
  } else {
    handleExportBookmarks()
  }
}

// 批量操作：表格选择变化
const handleSelectionChange = (rows: any[]) => {
  selectedBookmarkIds.value = rows.map((r) => r.id)
}

// 批量操作：取消全选
const clearSelection = () => {
  selectedBookmarkIds.value = []
}

// 批量操作：确认批量删除
const confirmBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedBookmarkIds.value.length} 个书签吗？此操作不可撤销。`,
      '确认批量删除',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )
    const result = await collectionStore.batchDeleteBookmarks(selectedBookmarkIds.value)
    if (result.success) {
      ElMessage.success('批量删除成功')
      selectedBookmarkIds.value = []
      if (managerCollectionId.value) {
        await collectionStore.fetchBookmarks({
          collectionId: managerCollectionId.value,
          folderId:
            managerFolderId.value === 'none' ? undefined : managerFolderId.value || undefined,
        })
      }
    } else {
      ElMessage.error(result.message || '批量删除失败')
    }
  } catch {
    // 用户取消
  }
}

// 批量操作：执行批量移动
const handleBatchMove = async () => {
  const result = await collectionStore.batchMoveBookmarks(
    selectedBookmarkIds.value,
    batchMoveFolderId.value || null,
  )
  if (result.success) {
    ElMessage.success('批量移动成功')
    showBatchMove.value = false
    batchMoveFolderId.value = ''
    selectedBookmarkIds.value = []
    if (managerCollectionId.value) {
      await collectionStore.fetchBookmarks({
        collectionId: managerCollectionId.value,
        folderId: managerFolderId.value === 'none' ? undefined : managerFolderId.value || undefined,
      })
    }
  } else {
    ElMessage.error(result.message || '批量移动失败')
  }
}
</script>

<style scoped>
.bookmark-manage-view {
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

.bookmark-table-wrapper {
  flex: 1;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
  overflow: auto;
  padding: 12px;
  margin-top: 4px;
  max-height: calc(100vh - 200px);
}

.batch-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  margin-bottom: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  color: #fff;
}

.batch-count {
  font-size: 14px;
  margin-right: 8px;
}

.batch-count strong {
  font-size: 16px;
}

.batch-bar .el-button {
  border-color: rgba(255, 255, 255, 0.4);
  color: #fff;
  background: rgba(255, 255, 255, 0.15);
}

.batch-bar .el-button:hover {
  background: rgba(255, 255, 255, 0.25);
}

.batch-bar .el-button--danger {
  background: rgba(255, 77, 79, 0.8);
  border-color: rgba(255, 77, 79, 0.6);
}

.batch-bar .el-button--danger:hover {
  background: rgba(255, 77, 79, 1);
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

.bookmark-title-cell {
  font-weight: 500;
  color: #333;
}

.bookmark-favicon {
  width: 20px;
  height: 20px;
  border-radius: 4px;
  object-fit: contain;
  vertical-align: middle;
}

.bookmark-icon-fallback {
  font-size: 20px;
}

.fallback-hidden {
  display: none;
}

/* 对话框和表单样式 */
.form-section {
  background: white;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 8px;
  margin-top: 4px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.section-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 15px;
  font-weight: 600;
  color: #333;
  margin-bottom: 8px;
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

.preview-icon-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
  border-radius: 8px;
  position: absolute;
  top: 0;
  left: 0;
}

.collection-management {
  padding: 0;
}

.collection-actions {
  margin-bottom: 12px;
  display: flex;
  justify-content: flex-end;
}

.collection-icon {
  font-size: 18px;
  display: inline-block;
}

.collection-name {
  font-weight: 600;
}

.collection-description {
  color: #666;
  font-size: 13px;
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
