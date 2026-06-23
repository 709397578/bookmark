<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import { isMobile } from '@/utils/device'
import DesktopLayout from '@/components/DesktopLayout.vue'
import AppLayout from '@/components/AppLayout.vue'

const route = useRoute()

// 根据设备类型选择布局组件
const LayoutComponent = computed(() => {
  // 登录页不需要布局
  if (route.path === '/login') {
    return null
  }

  // 根据设备类型返回不同的布局组件
  return isMobile() ? AppLayout : DesktopLayout
})
</script>

<template>
  <component :is="LayoutComponent" v-if="LayoutComponent">
    <RouterView />
  </component>
  <RouterView v-else />
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

#app {
  font-family:
    -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, 'Noto Sans',
    sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
</style>
