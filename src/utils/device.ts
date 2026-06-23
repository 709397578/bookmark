/**
 * 设备检测工具
 */

// 检测设备类型
export function isMobile(): boolean {
  // 通过屏幕宽度判断（768px 为分界点）
  return window.innerWidth < 768
}

// 获取当前设备类型
export function getDeviceType(): 'mobile' | 'desktop' {
  return isMobile() ? 'mobile' : 'desktop'
}

// 监听窗口大小变化
export function onResize(callback: (type: 'mobile' | 'desktop') => void) {
  let timeoutId: number

  window.addEventListener('resize', () => {
    // 防抖处理
    clearTimeout(timeoutId)
    timeoutId = window.setTimeout(() => {
      callback(getDeviceType())
    }, 300) as unknown as number
  })

  // 返回清理函数
  return () => {
    window.removeEventListener('resize', () => {})
    clearTimeout(timeoutId)
  }
}
