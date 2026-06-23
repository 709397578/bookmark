import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

// 请求拦截器 - 添加token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  },
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    if (error.response?.status === 401) {
      // 登录/注册接口的401是凭证错误，不是会话过期，不跳转
      const url = error.config?.url || ''
      if (!url.startsWith('/auth/')) {
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  },
)

// 认证相关API
export const authAPI = {
  login(data: { email: string; password: string }) {
    console.log('[authAPI.login] 发送登录请求', { email: data.email })
    return api.post('/auth/login', data)
  },
  register(data: { email: string; password: string; name?: string }) {
    console.log('[authAPI.register] 发送注册请求', { email: data.email })
    return api.post('/auth/register', data)
  },
  getProfile() {
    return api.get('/auth/profile')
  },
  listUsers() {
    return api.get('/auth/users')
  },
  updateUser(id: string, data: { name?: string; role?: string }) {
    return api.put(`/auth/users/${id}`, data)
  },
  deleteUser(id: string) {
    return api.delete(`/auth/users/${id}`)
  },
  uploadAvatar(file: File) {
    const formData = new FormData()
    formData.append('avatar', file)
    return api.post('/auth/avatar', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  deleteAvatar() {
    return api.delete('/auth/avatar')
  },
  changePassword(data: { oldPassword: string; newPassword: string }) {
    return api.put('/auth/password', data)
  },
}

// 收藏夹相关API
export const collectionAPI = {
  getCollections(publicOnly?: boolean) {
    return api.get('/collections', { params: { publicOnly } })
  },
  getCollectionById(id: string) {
    return api.get(`/collections/${id}`)
  },
  getCollectionBySlug(slug: string) {
    return api.get(`/collections/slug/${slug}`)
  },
  createCollection(data: {
    name: string
    description?: string
    icon?: string
    isPublic?: boolean
    sortOrder?: number
  }) {
    return api.post('/collections', data)
  },
  updateCollection(
    id: string,
    data: {
      name?: string
      description?: string
      icon?: string
      isPublic?: boolean
      sortOrder?: number
    },
  ) {
    return api.put(`/collections/${id}`, data)
  },
  deleteCollection(id: string) {
    return api.delete(`/collections/${id}`)
  },
  batchUpdateSortOrders(orders: { id: string; order: number }[]) {
    return api.put('/collections/batch/sort', { orders })
  },
}

// 文件夹相关API
export const folderAPI = {
  getFolders(collectionId?: string) {
    return api.get('/folders', { params: { collectionId } })
  },
  getFolderById(id: string) {
    return api.get(`/folders/${id}`)
  },
  createFolder(data: {
    name: string
    collectionId: string
    parentId?: string | null
    sortOrder?: number
    color?: string
  }) {
    return api.post('/folders', data)
  },
  updateFolder(
    id: string,
    data: {
      name?: string
      icon?: string
      parentId?: string | null
      sortOrder?: number
      color?: string
    },
  ) {
    return api.put(`/folders/${id}`, data)
  },
  deleteFolder(id: string) {
    return api.delete(`/folders/${id}`)
  },
  batchUpdateSortOrders(orders: { id: string; order: number }[]) {
    return api.put('/folders/batch/sort', { orders })
  },
}

// 书签相关API
export const bookmarkAPI = {
  getBookmarks(params: {
    page?: number
    pageSize?: number
    collectionId?: string
    folderId?: string
  }) {
    return api.get('/bookmarks', { params })
  },
  getBookmarkById(id: string) {
    return api.get(`/bookmarks/${id}`)
  },
  createBookmark(data: {
    title: string
    url: string
    description?: string
    icon?: string
    collectionId: string
    folderId?: string | null
    tags?: string[]
    isFeatured?: boolean
    sortOrder?: number
    hasSnapshot?: boolean
    snapshotUrl?: string
  }) {
    return api.post('/bookmarks', data)
  },
  updateBookmark(
    id: string,
    data: {
      title?: string
      url?: string
      description?: string
      icon?: string
      folderId?: string | null
      isFeatured?: boolean
      sortOrder?: number
      hasSnapshot?: boolean
      snapshotUrl?: string
    },
  ) {
    return api.put(`/bookmarks/${id}`, data)
  },
  deleteBookmark(id: string) {
    return api.delete(`/bookmarks/${id}`)
  },
  searchBookmarks(query: string) {
    return api.get('/bookmarks/search', { params: { q: query } })
  },
  generateSnapshot(id: string) {
    return api.post(`/bookmarks/${id}/snapshot`)
  },
  exportBookmarks(collectionId: string) {
    return api.get('/bookmarks/export', { params: { collectionId } })
  },
  exportBookmarksHTML(collectionId: string) {
    return api.get('/bookmarks/export/html', {
      params: { collectionId },
      responseType: 'blob',
    })
  },
  importBookmarksHTML(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/bookmarks/import/html', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  batchDelete(ids: string[]) {
    return api.post('/bookmarks/batch/delete', { ids })
  },
  batchMove(ids: string[], folderId: string | null) {
    return api.put('/bookmarks/batch/move', { ids, folderId })
  },
  batchUpdateSortOrders(orders: { id: string; order: number }[]) {
    return api.put('/bookmarks/batch/sort', { orders })
  },
}

// 设置相关API
export const settingAPI = {
  getSettings() {
    return api.get('/settings')
  },
  getPublicSettings() {
    return api.get('/settings/public')
  },
  getSettingByKey(key: string) {
    return api.get(`/settings/${key}`)
  },
  updateSetting(key: string, value: any) {
    return api.put(`/settings/${key}`, { value })
  },
}

export default api
