<template>
  <div class="home-view">
    <!-- 空状态 -->
    <van-empty
      v-if="!collectionStore.currentCollection && searchResults.length === 0"
      description="请先选择一个收藏夹"
      image="search"
    />

    <!-- 书签列表 -->
    <div
      v-if="collectionStore.currentCollection || searchResults.length > 0"
      class="bookmarks-container"
    >
      <van-skeleton v-if="collectionStore.loading" :rows="5" animated />

      <van-empty
        v-else-if="displayBookmarks.length === 0"
        :description="!!searchQuery ? '未找到匹配的书签' : '暂无书签，点击右下角添加'"
        image="search"
      />

      <div v-else class="bookmarks-grid">
        <div
          v-for="bookmark in displayBookmarks"
          :key="bookmark.id"
          class="bookmark-card"
          :class="{ featured: bookmark.isFeatured }"
        >
          <div class="bookmark-content">
            <div class="bookmark-left">
              <span class="bookmark-icon">
                <img
                  v-if="getFaviconUrl(bookmark.url)"
                  :src="getFaviconUrl(bookmark.url)"
                  :alt="bookmark.title"
                  class="bookmark-favicon"
                  @error="
                    ;($event.target as HTMLImageElement).style.display = 'none'
                    ;($event.target as HTMLImageElement).nextElementSibling?.classList.remove(
                      'fallback-hidden',
                    )
                  "
                />
                <span
                  :class="[
                    'bookmark-icon-fallback',
                    { 'fallback-hidden': getFaviconUrl(bookmark.url) },
                  ]"
                  >{{ bookmark.icon || '🔗' }}</span
                >
              </span>
              <div
                class="snapshot-indicator"
                v-if="bookmark.hasSnapshot"
                @click="openSnapshot(bookmark)"
              >
                <van-icon name="photograph" />
              </div>
            </div>
            <div class="bookmark-right">
              <h3 class="bookmark-title" @click="openBookmark(bookmark)">
                {{ bookmark.title }}
              </h3>
              <p class="bookmark-url" @click="openBookmark(bookmark)">
                {{ truncateUrl(bookmark.url) }}
              </p>
              <p v-if="bookmark.description" class="bookmark-description">
                {{ bookmark.description }}
              </p>
              <div class="bookmark-footer">
                <span class="bookmark-time">{{ formatTime(bookmark.updatedAt) }}</span>
                <van-tag v-if="bookmark.isFeatured" type="danger">精选</van-tag>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, computed, watch } from 'vue'
import { useCollectionStore } from '@/stores/collection'
import { getFaviconUrl } from '@/utils/ui'
import { useRouter, useRoute } from 'vue-router'
import { showToast } from 'vant'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const collectionStore = useCollectionStore()
const router = useRouter()
const route = useRoute()

// 搜索相关状态
const searchQuery = ref('')
const searchResults = ref<any[]>([])

// 计算当前显示的书签列表
const displayBookmarks = computed(() => {
  return searchResults.value.length > 0 ? searchResults.value : collectionStore.bookmarks
})

// 计算当前标题
const displayTitle = computed(() => {
  return searchQuery.value
    ? `搜索: ${searchQuery.value}`
    : collectionStore.currentCollection?.name || '书签'
})

// 初始化
onMounted(async () => {
  // 根据登录状态决定是否只获取公开收藏夹
  const publicOnly = !authStore.isAuthenticated
  await collectionStore.fetchCollections(publicOnly)

  // 如果没有当前收藏夹，自动选择第一个
  if (!collectionStore.currentCollection && collectionStore.collections.length > 0) {
    // 找到第一个用户有权限访问的收藏夹
    const accessibleCollection = collectionStore.collections.find(
      (c) => c.isPublic || authStore.isAuthenticated,
    )

    if (accessibleCollection) {
      collectionStore.setCurrentCollection(accessibleCollection)
      await collectionStore.fetchFolders(accessibleCollection.id)
      await collectionStore.fetchBookmarks({ collectionId: accessibleCollection.id })
    }
  }

  // 监听搜索结果更新事件
  window.addEventListener('searchResultsUpdated', (event: any) => {
    const { query, results } = event.detail
    searchQuery.value = query
    searchResults.value = results

    showToast(`找到 ${results.length} 个结果`)
  })

  // 监听清除搜索结果事件
  window.addEventListener('clearSearchResults', () => {
    searchQuery.value = ''
    searchResults.value = []
  })
})

// 监听路由变化，清除搜索状态
watch(
  () => route.path,
  () => {
    searchQuery.value = ''
    searchResults.value = []
  },
)

// 打开书签
const openBookmark = (bookmark: any) => {
  window.open(bookmark.url, '_blank')
}

// 打开快照
const openSnapshot = (bookmark: any) => {
  if (bookmark.snapshotUrl) {
    window.open(bookmark.snapshotUrl, '_blank')
  } else {
    showToast('快照文件不存在')
  }
}

// 截断URL显示
const truncateUrl = (url: string) => {
  try {
    const urlObj = new URL(url)
    return urlObj.hostname + urlObj.pathname
  } catch {
    return url.length > 50 ? url.substring(0, 50) + '...' : url
  }
}

// 格式化时间
const formatTime = (timeStr: string) => {
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`

  return date.toLocaleDateString('zh-CN')
}
</script>

<style scoped>
.home-view {
  height: 100%;
}

.bookmarks-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.bookmarks-header {
  margin-bottom: 24px;
  padding: 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.bookmarks-header h2 {
  margin: 0 0 8px 0;
  font-size: 28px;
  color: #333;
  font-weight: 600;
}

.description {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bookmarks-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
  padding: 4px;
}

.bookmark-card {
  background: white;
  border-radius: 10px;
  padding: 16px;
  transition: all 0.25s ease;
  cursor: pointer;
  border: 1px solid #f0f0f0;
  position: relative;
  overflow: hidden;
  animation: fadeInUp 0.4s ease-out;
}

.bookmark-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  transform: scaleX(0);
  transition: transform 0.3s ease;
}

.bookmark-card:hover::before {
  transform: scaleX(1);
}

.bookmark-card:hover {
  border-color: #e0e0e0;
  background: #fafafa;
}

.bookmark-card.featured {
  border: 1px solid #ffccc7;
  background: #fff5f5;
}

.bookmark-card.featured::before {
  background: linear-gradient(90deg, #ff4d4f 0%, #ff7875 100%);
}

.bookmark-content {
  display: flex;
  align-items: flex-start;
}

.bookmark-left {
  flex-shrink: 0;
  margin-right: 16px;
  margin-top: 4px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.snapshot-indicator {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #667eea;
  font-size: 16px;
  cursor: pointer;
  transition:
    transform 0.2s ease,
    color 0.2s ease;
  border-radius: 4px;
  padding: 2px;
}

.snapshot-indicator:hover {
  transform: scale(1.2);
  color: #5a67d8;
  background: rgba(102, 126, 234, 0.1);
}

.snapshot-indicator:active {
  transform: scale(0.95);
}

.bookmark-right {
  flex: 1;
  min-width: 0;
}

.bookmark-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
}

.bookmark-favicon {
  width: 24px;
  height: 24px;
  border-radius: 4px;
  object-fit: contain;
}

.bookmark-icon-fallback {
  font-size: 24px;
  line-height: 1;
}

.fallback-hidden {
  display: none;
}

.bookmark-title {
  margin: 0 0 10px 0;
  font-size: 17px;
  color: #333;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.4;
}

.bookmark-title:hover {
  color: #667eea;
}

.bookmark-url {
  margin: 0 0 10px 0;
  font-size: 13px;
  color: #999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
  font-family: 'Monaco', 'Consolas', monospace;
}

.bookmark-url:hover {
  color: #667eea;
}

.bookmark-description {
  margin: 0 0 14px 0;
  font-size: 14px;
  color: #666;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  line-clamp: 2;
  overflow: hidden;
}

.bookmark-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 14px;
  border-top: 1px solid #f0f0f0;
}

.bookmark-time {
  font-size: 12px;
  color: #999;
}

/* 响应式设计 - 桌面端优化 */
@media (min-width: 1920px) {
  .bookmarks-grid {
    grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
    gap: 24px;
  }

  .bookmark-card {
    padding: 24px;
  }

  .bookmarks-header h2 {
    font-size: 32px;
  }
}

@media (min-width: 2560px) {
  .bookmarks-grid {
    grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
    gap: 28px;
  }

  .bookmark-card {
    padding: 28px;
  }

  .bookmark-title {
    font-size: 19px;
  }

  .bookmark-description {
    font-size: 15px;
  }
}

@media (max-width: 1440px) {
  .bookmarks-grid {
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
  }
}

@media (max-width: 1024px) {
  .bookmarks-grid {
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 14px;
  }

  .bookmark-card {
    padding: 16px;
  }
}

/* 动画效果 */
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 手机端适配 */
@media (max-width: 768px) {
  .bookmarks-grid {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .bookmark-card {
    padding: 14px;
    border-radius: 14px;
  }

  .bookmark-title {
    font-size: 15px;
    margin-bottom: 6px;
  }

  .bookmark-url {
    font-size: 12px;
    margin-bottom: 8px;
  }

  .bookmark-description {
    font-size: 13px;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    margin-bottom: 10px;
  }

  .bookmark-left {
    margin-right: 12px;
  }

  .bookmark-icon {
    width: 28px;
    height: 28px;
  }

  .bookmark-icon-fallback {
    font-size: 20px;
  }

  .bookmark-favicon {
    width: 20px;
    height: 20px;
  }

  .bookmark-footer {
    padding-top: 10px;
  }
}

@media (max-width: 375px) {
  .bookmarks-grid {
    gap: 10px;
  }

  .bookmark-card {
    padding: 12px;
    border-radius: 12px;
  }
}

.bookmark-card:nth-child(1) {
  animation-delay: 0.05s;
}
.bookmark-card:nth-child(2) {
  animation-delay: 0.1s;
}
.bookmark-card:nth-child(3) {
  animation-delay: 0.15s;
}
.bookmark-card:nth-child(4) {
  animation-delay: 0.2s;
}

.bookmark-card:nth-child(5) {
  animation-delay: 0.25s;
}
.bookmark-card:nth-child(6) {
  animation-delay: 0.3s;
}
</style>
