import { ref } from 'vue'
import { defineStore } from 'pinia'
import { collectionAPI, folderAPI, bookmarkAPI } from '@/api'

// 统一的响应检查函数
function isSuccessResponse(response: any): boolean {
  // 支持两种格式：{ code: 200, ... } 和 { success: true, ... }
  return response?.code === 200 || response?.success === true
}

function getResponseData(response: any): any {
  return response?.data
}

function getResponseMessage(response: any): string {
  return response?.message || response?.error || ''
}

export interface Collection {
  id: string
  name: string
  slug: string
  description?: string
  icon?: string
  isPublic: boolean
  sortOrder: number
  userId: string
  createdAt: string
  updatedAt: string
}

export interface Folder {
  id: string
  name: string
  icon?: string
  color?: string
  collectionId: string
  parentId?: string | null
  sortOrder: number
  createdAt: string
  updatedAt: string
}

export interface Bookmark {
  id: string
  title: string
  url: string
  description?: string
  icon?: string
  isFeatured: boolean
  sortOrder: number
  collectionId: string
  folderId?: string | null
  tags?: any[]
  hasSnapshot?: boolean
  snapshotUrl?: string
  createdAt: string
  updatedAt: string
}

export const useCollectionStore = defineStore('collection', () => {
  const collections = ref<Collection[]>([])
  const currentCollection = ref<Collection | null>(null)
  const folders = ref<Folder[]>([])
  const bookmarks = ref<Bookmark[]>([])
  const loading = ref(false)

  // 获取收藏夹列表
  async function fetchCollections(publicOnly?: boolean) {
    loading.value = true
    try {
      const response: any = await collectionAPI.getCollections(publicOnly)
      if (isSuccessResponse(response)) {
        collections.value = getResponseData(response)
        return getResponseData(response)
      }
    } catch (error) {
      console.error('获取收藏夹失败:', error)
    } finally {
      loading.value = false
    }
  }

  // 设置当前收藏夹
  function setCurrentCollection(collection: Collection | null) {
    currentCollection.value = collection
  }

  // 更新收藏夹
  async function updateCollection(
    id: string,
    data: {
      name?: string
      description?: string
      icon?: string
      isPublic?: boolean
      sortOrder?: number
    },
  ) {
    try {
      const response: any = await collectionAPI.updateCollection(id, data)
      if (isSuccessResponse(response)) {
        const index = collections.value.findIndex((c) => c.id === id)
        if (index !== -1) {
          collections.value[index] = { ...collections.value[index], ...data } as Collection
        }
        if (currentCollection.value?.id === id) {
          currentCollection.value = { ...currentCollection.value, ...data }
        }
        return { success: true }
      }
      return { success: false, message: getResponseMessage(response) }
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || '更新失败',
      }
    }
  }

  // 获取文件夹列表
  async function fetchFolders(collectionId?: string) {
    loading.value = true
    try {
      const response: any = await folderAPI.getFolders(collectionId)
      if (isSuccessResponse(response)) {
        folders.value = getResponseData(response)
        return getResponseData(response)
      }
    } catch (error) {
      console.error('获取文件夹失败:', error)
    } finally {
      loading.value = false
    }
  }

  // 获取书签列表
  async function fetchBookmarks(params: {
    page?: number
    pageSize?: number
    collectionId?: string
    folderId?: string
  }) {
    loading.value = true
    try {
      const response: any = await bookmarkAPI.getBookmarks(params)
      if (isSuccessResponse(response)) {
        const data = getResponseData(response)
        bookmarks.value = data.data || data
        return data
      }
    } catch (error) {
      console.error('获取书签失败:', error)
    } finally {
      loading.value = false
    }
  }

  // 创建收藏夹
  async function createCollection(data: {
    name: string
    description?: string
    icon?: string
    isPublic?: boolean
  }) {
    try {
      const response: any = await collectionAPI.createCollection(data)
      if (isSuccessResponse(response)) {
        const responseData = getResponseData(response)
        collections.value.push(responseData)
        return { success: true, data: responseData }
      }
      return { success: false, message: getResponseMessage(response) }
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || '创建失败',
      }
    }
  }

  // 删除收藏夹
  async function deleteCollection(id: string) {
    try {
      const response: any = await collectionAPI.deleteCollection(id)
      if (isSuccessResponse(response)) {
        collections.value = collections.value.filter((c) => c.id !== id)
        if (currentCollection.value?.id === id) {
          currentCollection.value = null
        }
        return { success: true }
      }
      return { success: false, message: getResponseMessage(response) }
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || '删除失败',
      }
    }
  }

  // 批量更新收藏夹排序
  async function batchUpdateCollectionSortOrders(orders: { id: string; order: number }[]) {
    try {
      const response: any = await collectionAPI.batchUpdateSortOrders(orders)
      if (isSuccessResponse(response)) {
        orders.forEach(({ id, order }) => {
          const col = collections.value.find((c) => c.id === id)
          if (col) col.sortOrder = order
        })
        return { success: true }
      }
      return { success: false, message: getResponseMessage(response) }
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || '更新排序失败',
      }
    }
  }

  // 创建文件夹
  async function createFolder(data: {
    name: string
    collectionId: string
    icon?: string
    color?: string
    parentId?: string | null
  }) {
    try {
      const response: any = await folderAPI.createFolder(data)
      if (isSuccessResponse(response)) {
        const responseData = getResponseData(response)
        folders.value.push(responseData)
        return { success: true, data: responseData }
      }
      return { success: false, message: getResponseMessage(response) }
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || '创建失败',
      }
    }
  }

  // 删除文件夹
  async function deleteFolder(id: string) {
    try {
      const response: any = await folderAPI.deleteFolder(id)
      if (isSuccessResponse(response)) {
        folders.value = folders.value.filter((f) => f.id !== id)
        return { success: true }
      }
      return { success: false, message: getResponseMessage(response) }
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || '删除失败',
      }
    }
  }

  // 批量更新文件夹排序
  async function batchUpdateFolderSortOrders(orders: { id: string; order: number }[]) {
    try {
      const response: any = await folderAPI.batchUpdateSortOrders(orders)
      if (isSuccessResponse(response)) {
        orders.forEach(({ id, order }) => {
          const f = folders.value.find((f) => f.id === id)
          if (f) f.sortOrder = order
        })
        return { success: true }
      }
      return { success: false, message: getResponseMessage(response) }
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || '更新排序失败',
      }
    }
  }

  // 更新文件夹
  async function updateFolder(
    id: string,
    data: {
      name?: string
      icon?: string
      color?: string
    },
  ) {
    try {
      const response: any = await folderAPI.updateFolder(id, data)
      if (isSuccessResponse(response)) {
        const index = folders.value.findIndex((f) => f.id === id)
        if (index !== -1) {
          folders.value[index] = { ...folders.value[index], ...data } as Folder
        }
        return { success: true }
      }
      return { success: false, message: getResponseMessage(response) }
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || '更新失败',
      }
    }
  }

  // 创建书签
  async function createBookmark(data: {
    title: string
    url: string
    description?: string
    icon?: string
    collectionId: string
    folderId?: string | null
    isFeatured?: boolean
    hasSnapshot?: boolean
    snapshotUrl?: string
  }) {
    try {
      const response: any = await bookmarkAPI.createBookmark(data)
      if (isSuccessResponse(response)) {
        const responseData = getResponseData(response)
        bookmarks.value.unshift(responseData)
        return { success: true, data: responseData }
      }
      return { success: false, message: getResponseMessage(response) }
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || '创建失败',
      }
    }
  }

  // 更新书签
  async function updateBookmark(
    id: string,
    data: {
      title?: string
      url?: string
      description?: string
      icon?: string
      folderId?: string | null
      isFeatured?: boolean
      hasSnapshot?: boolean
      snapshotUrl?: string
    },
  ) {
    try {
      const response: any = await bookmarkAPI.updateBookmark(id, data)
      if (isSuccessResponse(response)) {
        const responseData = getResponseData(response)
        const index = bookmarks.value.findIndex((b) => b.id === id)
        if (index !== -1) {
          bookmarks.value[index] = responseData
        }
        return { success: true, data: responseData }
      }
      return { success: false, message: getResponseMessage(response) }
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || '更新失败',
      }
    }
  }

  // 删除书签
  async function deleteBookmark(id: string) {
    try {
      const response: any = await bookmarkAPI.deleteBookmark(id)
      if (isSuccessResponse(response)) {
        bookmarks.value = bookmarks.value.filter((b) => b.id !== id)
        return { success: true }
      }
      return { success: false, message: getResponseMessage(response) }
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || '删除失败',
      }
    }
  }

  // 搜索书签
  async function searchBookmarks(query: string) {
    loading.value = true
    try {
      const response: any = await bookmarkAPI.searchBookmarks(query)
      if (isSuccessResponse(response)) {
        return getResponseData(response)
      }
    } catch (error) {
      console.error('搜索失败:', error)
    } finally {
      loading.value = false
    }
  }

  // 批量删除书签
  async function batchDeleteBookmarks(ids: string[]) {
    try {
      const response: any = await bookmarkAPI.batchDelete(ids)
      if (isSuccessResponse(response)) {
        bookmarks.value = bookmarks.value.filter((b) => !ids.includes(b.id))
        return { success: true }
      }
      return { success: false, message: getResponseMessage(response) }
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || '批量删除失败',
      }
    }
  }

  // 批量移动书签
  async function batchMoveBookmarks(ids: string[], folderId: string | null) {
    try {
      const response: any = await bookmarkAPI.batchMove(ids, folderId)
      if (isSuccessResponse(response)) {
        bookmarks.value = bookmarks.value.map((b) =>
          ids.includes(b.id) ? { ...b, folderId: folderId || null } : b,
        )
        return { success: true }
      }
      return { success: false, message: getResponseMessage(response) }
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || '批量移动失败',
      }
    }
  }

  // 批量更新排序
  async function batchUpdateSortOrders(orders: { id: string; order: number }[]) {
    try {
      const response: any = await bookmarkAPI.batchUpdateSortOrders(orders)
      if (isSuccessResponse(response)) {
        orders.forEach(({ id, order }) => {
          const bm = bookmarks.value.find((b) => b.id === id)
          if (bm) bm.sortOrder = order
        })
        return { success: true }
      }
      return { success: false, message: getResponseMessage(response) }
    } catch (error: any) {
      return {
        success: false,
        message: error.response?.data?.message || '更新排序失败',
      }
    }
  }

  return {
    collections,
    currentCollection,
    folders,
    updateFolder,
    bookmarks,
    loading,
    fetchCollections,
    setCurrentCollection,
    updateCollection,
    fetchFolders,
    fetchBookmarks,
    createCollection,
    deleteCollection,
    createFolder,
    deleteFolder,
    createBookmark,
    updateBookmark,
    deleteBookmark,
    searchBookmarks,
    batchDeleteBookmarks,
    batchMoveBookmarks,
    batchUpdateSortOrders,
    batchUpdateCollectionSortOrders,
    batchUpdateFolderSortOrders,
  }
})
