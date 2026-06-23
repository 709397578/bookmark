/**
 * UI 组件适配器
 * 根据设备类型自动选择使用 Vant 或 Element Plus 组件
 */

import { isMobile } from './device'

// 消息提示组件
export function showToast(message: string, options?: any) {
  if (isMobile()) {
    // Vant Toast
    return import('vant').then(({ showNotify }) => {
      showNotify({ type: 'success', message, ...options })
    })
  } else {
    // Element Plus Message
    return import('element-plus').then(({ ElMessage }) => {
      ElMessage.success({ message, ...options })
    })
  }
}

// 确认对话框
export function showConfirmDialog(options: {
  title?: string
  message: string
  confirmButtonText?: string
  cancelButtonText?: string
}): Promise<boolean> {
  return new Promise((resolve) => {
    if (isMobile()) {
      // Vant Dialog
      import('vant').then(({ showDialog }) => {
        showDialog({
          title: options.title || '提示',
          message: options.message,
          confirmButtonText: options.confirmButtonText || '确认',
          cancelButtonText: options.cancelButtonText || '取消',
          showCancelButton: true,
        })
          .then(() => resolve(true))
          .catch(() => resolve(false))
      })
    } else {
      // Element Plus MessageBox
      import('element-plus').then(({ ElMessageBox }) => {
        ElMessageBox.confirm(options.message, options.title || '提示', {
          confirmButtonText: options.confirmButtonText || '确认',
          cancelButtonText: options.cancelButtonText || '取消',
          type: 'warning',
        })
          .then(() => resolve(true))
          .catch(() => resolve(false))
      })
    }
  })
}

// 加载状态
export function showLoading(message?: string) {
  if (isMobile()) {
    // Vant Loading
    return import('vant').then(({ showLoadingToast }) => {
      return showLoadingToast({
        message: message || '加载中...',
        forbidClick: true,
        duration: 0,
      })
    })
  } else {
    // Element Plus Loading
    return import('element-plus').then(({ ElLoading }) => {
      return ElLoading.service({
        lock: true,
        text: message || '加载中...',
        background: 'rgba(0, 0, 0, 0.7)',
      })
    })
  }
}

// 隐藏加载状态
export function hideLoading(instance?: any) {
  if (instance) {
    instance.close && instance.close()
  }
}

/**
 * 获取网站 favicon URL
 * 直接使用网站自身的 /favicon.ico，不依赖第三方服务
 */
export function getFaviconUrl(url: string, size: number = 32): string {
  try {
    const urlObj = new URL(url)
    return `${urlObj.origin}/favicon.ico`
  } catch {
    return ''
  }
}
