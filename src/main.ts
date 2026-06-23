import './assets/main.css'
import './assets/theme.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { isMobile } from '@/utils/device'

import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())
app.use(router)

// 根据设备类型注册 UI 框架（使用动态导入实现代码分割）
async function setupUIFramework() {
  if (isMobile()) {
    // 移动端使用 Vant
    const Vant = await import('vant')
    await import('vant/lib/index.css')
    app.use(Vant.default)
  } else {
    // 桌面端使用 Element Plus
    const ElementPlus = await import('element-plus')
    await import('element-plus/dist/index.css')
    app.use(ElementPlus.default)

    // 注册所有 Element Plus 图标
    const Icons = await import('@element-plus/icons-vue')
    for (const [key, component] of Object.entries(Icons)) {
      app.component(key, component)
    }
  }

  // UI 框架注册完成后挂载应用
  app.mount('#app')
}

// 执行初始化
setupUIFramework()
