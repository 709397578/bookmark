<template>
  <div class="bookmark-import-view">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <el-button :icon="ArrowLeft" @click="$router.push('/bookmarks')">返回书签管理</el-button>
        <h2 class="page-title">
          <el-icon><Upload /></el-icon>
          导入书签
        </h2>
      </div>
    </div>

    <!-- 导入区域 -->
    <div class="import-section">
      <el-card class="upload-card">
        <template #header>
          <div class="card-header">
            <span>选择书签文件</span>
            <div>
              <el-tag type="info" size="small" style="margin-right: 6px">JSON</el-tag>
              <el-tag type="success" size="small">Chrome HTML</el-tag>
            </div>
          </div>
        </template>

        <div class="upload-area">
          <el-upload
            ref="uploadRef"
            class="bookmark-uploader"
            drag
            :auto-upload="false"
            :show-file-list="true"
            :limit="1"
            accept=".json,.html"
            :on-change="handleFileChange"
          >
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
            <template #tip>
              <div class="el-upload__tip">
                支持 JSON 和 Chrome HTML (Netscape) 格式的书签文件，文件大小不超过 10MB
              </div>
            </template>
          </el-upload>
        </div>
      </el-card>

      <!-- 预览区域 -->
      <el-card v-if="previewData" class="preview-card">
        <template #header>
          <div class="card-header">
            <span>书签预览</span>
            <el-tag type="success" size="small">共 {{ totalBookmarks }} 个书签</el-tag>
          </div>
        </template>

        <div class="preview-controls">
          <el-select
            v-model="selectedCollectionId"
            placeholder="选择目标收藏夹"
            style="width: 250px"
            @change="handleCollectionChange"
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

          <el-button
            type="primary"
            :icon="Upload"
            :loading="importing"
            @click="handleImport"
            :disabled="!selectedCollectionId || importing"
          >
            {{ importing ? '导入中...' : '导入书签' }}
          </el-button>
        </div>

        <div class="bookmark-preview">
          <el-collapse v-model="activeCollapse">
            <el-collapse-item v-for="(folder, index) in previewData" :key="index" :name="index">
              <template #title>
                <span class="folder-title">{{ folder.title || `文件夹 ${index + 1}` }}</span>
                <el-tag type="info" size="small" v-if="folder.children">
                  {{ folder.children.length }} 个项目
                </el-tag>
              </template>

              <div v-if="folder.children && folder.children.length > 0" class="bookmark-list">
                <div
                  v-for="(bookmark, bIndex) in folder.children"
                  :key="bIndex"
                  class="bookmark-item"
                >
                  <div class="bookmark-info">
                    <div class="bookmark-title">{{ bookmark.title }}</div>
                    <div class="bookmark-url">{{ bookmark.url }}</div>
                  </div>
                </div>
              </div>
            </el-collapse-item>
          </el-collapse>

          <el-empty
            v-if="!previewData || previewData.length === 0"
            description="没有可预览的书签"
          />
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadInstance, UploadProps, UploadRawFile } from 'element-plus'
import { ArrowLeft, Upload, UploadFilled } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { useCollectionStore } from '@/stores/collection'
import { bookmarkAPI } from '@/api'

const router = useRouter()
const authStore = useAuthStore()
const collectionStore = useCollectionStore()

const uploadRef = ref<UploadInstance>()
const previewData = ref<any[]>([])
const selectedCollectionId = ref<string>('')
const importing = ref(false)
const activeCollapse = ref<string[]>([])

// 计算总书签数
const totalBookmarks = computed(() => {
  if (!previewData.value) return 0

  let count = 0
  const countBookmarks = (items: any[]) => {
    items.forEach((item) => {
      if (item.type === 'link') {
        count++
      } else if (item.children) {
        countBookmarks(item.children)
      }
    })
  }

  countBookmarks(previewData.value)

  return count
})

// 初始化
onMounted(async () => {
  await collectionStore.fetchCollections()
  // 如果有收藏夹，自动选择第一个
  if (collectionStore.collections.length > 0) {
    selectedCollectionId.value = collectionStore.collections[0]!.id
  }
})

// 处理文件选择变化
const handleFileChange: UploadProps['onChange'] = (uploadFile) => {
  const file = uploadFile.raw as UploadRawFile
  if (!file) return

  const isJSON = file.type === 'application/json' || file.name.endsWith('.json')
  const isHTML = file.type === 'text/html' || file.name.endsWith('.html')

  if (!isJSON && !isHTML) {
    ElMessage.error('请上传 JSON 或 HTML 格式的书签文件')
    return
  }

  // 检查文件大小 (10MB)
  const isLt10M = file.size / 1024 / 1024 < 10
  if (!isLt10M) {
    ElMessage.error('文件大小不能超过 10MB')
    return
  }

  if (isHTML) {
    handleHTMLFile(file)
  } else {
    handleJSONFile(file)
  }
}

// 处理 HTML 格式书签文件
const handleHTMLFile = async (file: UploadRawFile) => {
  try {
    const response = await bookmarkAPI.importBookmarksHTML(file)
    const data = response.data

    if (Array.isArray(data)) {
      const processedData = processData(data)
      previewData.value = processedData
      ElMessage.success('HTML 书签文件解析成功')
      activeCollapse.value = previewData.value.map((_: any, index: number) => index.toString())
    } else {
      ElMessage.error('解析书签文件失败')
    }
  } catch (error) {
    console.error('解析HTML书签文件失败:', error)
    ElMessage.error('解析HTML书签文件失败，请检查文件格式')
  }
}

// 处理 JSON 格式书签文件
const handleJSONFile = (file: UploadRawFile) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const content = e.target?.result as string
      const data = JSON.parse(content)

      if (!Array.isArray(data)) {
        ElMessage.error('无效的书签文件格式')
        return
      }

      const processedData = processData(data)
      previewData.value = processedData
      ElMessage.success('文件解析成功')
      activeCollapse.value = previewData.value.map((_: any, index: number) => index.toString())
    } catch (error) {
      console.error('解析书签文件失败:', error)
      ElMessage.error('解析书签文件失败，请检查文件格式')
    }
  }
  reader.readAsText(file)
}

// 处理导入的书签数据，移除icon字段
const processData = (data: any[]): any[] => {
  return data.map((item) => {
    if (item.type === 'link') {
      // 移除icon字段，因为系统有处理icon的函数
      const { icon, ...rest } = item
      return rest
    } else if (item.type === 'folder' && item.children) {
      // 递归处理子项
      return {
        ...item,
        children: processData(item.children),
      }
    }
    return item
  })
}

// 收藏夹切换处理
const handleCollectionChange = (collectionId: string) => {
  selectedCollectionId.value = collectionId
}

// 处理导入
const handleImport = async () => {
  if (!selectedCollectionId.value) {
    ElMessage.warning('请选择目标收藏夹')
    return
  }

  try {
    // 确认导入
    await ElMessageBox.confirm(
      `确定要将 ${totalBookmarks.value} 个书签导入到选中的收藏夹吗？`,
      '确认导入',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      },
    )

    importing.value = true

    // 递归处理文件夹和书签
    const processItems = async (items: any[], parentId: string | null = null) => {
      let folderCreated = false
      let currentFolderId = parentId

      for (const item of items) {
        if (item.type === 'folder') {
          // 创建文件夹
          const folderData = {
            name: item.title || '未命名文件夹',
            collectionId: selectedCollectionId.value,
            parentId: currentFolderId,
          }

          try {
            const folderResult = await collectionStore.createFolder(folderData)
            if (folderResult.success) {
              folderCreated = true
              currentFolderId = folderResult.data.id

              // 递归处理子项
              if (item.children && item.children.length > 0) {
                await processItems(item.children, currentFolderId)
              }
            } else {
              console.error('创建文件夹失败:', folderResult.message)
            }
          } catch (error) {
            console.error('创建文件夹异常:', error)
          }
        } else if (item.type === 'link' && item.url) {
          // 创建书签
          const bookmarkData = {
            title: item.title,
            url: item.url,
            description: item.description || '',
            collectionId: selectedCollectionId.value,
            folderId: currentFolderId,
            sortOrder: item.sortOrder || 0,
          }

          try {
            await bookmarkAPI.createBookmark(bookmarkData)
          } catch (error) {
            console.error('创建书签失败:', error)
          }
        }
      }
    }

    // 开始处理
    await processItems(previewData.value)

    ElMessage.success('导入成功')
    // 返回书签管理页面
    router.push('/bookmarks')
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('导入失败:', error)
      ElMessage.error('导入失败，请重试')
    }
  } finally {
    importing.value = false
  }
}
</script>

<style scoped>
.bookmark-import-view {
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

.import-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 12px;
}

.upload-card,
.preview-card {
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.upload-area {
  display: flex;
  justify-content: center;
  margin: 20px 0;
}

.bookmark-uploader {
  width: 100%;
  max-width: 600px;
}

.preview-controls {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}

.bookmark-preview {
  max-height: 500px;
  overflow-y: auto;
}

.bookmark-list {
  padding: 8px 0;
}

.bookmark-item {
  display: flex;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
}

.bookmark-item:last-child {
  border-bottom: none;
}

.bookmark-info {
  flex: 1;
}

.bookmark-title {
  font-weight: 500;
  margin-bottom: 4px;
  color: #333;
}

.bookmark-url {
  font-size: 14px;
  color: #666;
  word-break: break-all;
}

.folder-title {
  font-weight: 600;
  color: #333;
}

/* 响应式 */
@media (max-width: 768px) {
  .preview-controls {
    flex-direction: column;
    align-items: stretch;
  }

  .preview-controls .el-select {
    width: 100%;
  }
}
</style>
