<template>
  <div class="about-view">
    <div class="about-container">
      <h1>关于 PinTree</h1>

      <div class="info-card">
        <h2>📚 项目介绍</h2>
        <p>PinTree 是一个现代化的书签管理系统，帮助您高效地组织和管理网页收藏。</p>
      </div>

      <div class="info-card">
        <h2>✨ 主要功能</h2>
        <ul>
          <li>✅ 收藏夹管理 - 创建和管理多个收藏夹</li>
          <li>✅ 文件夹分类 - 在收藏夹内创建文件夹进行分类</li>
          <li>✅ 书签管理 - 添加、编辑、删除书签</li>
          <li>✅ 全局搜索 - 快速搜索所有书签</li>
          <li>✅ 精选标记 - 标记重要书签</li>
          <li>✅ 响应式设计 - 支持各种设备访问</li>
        </ul>
      </div>

      <div class="info-card">
        <h2>🛠️ 技术栈</h2>
        <div class="tech-stack">
          <span class="tech-tag">Vue 3</span>
          <span class="tech-tag">TypeScript</span>
          <span class="tech-tag">Vite</span>
          <span class="tech-tag">Pinia</span>
          <span class="tech-tag">Vue Router</span>
          <span class="tech-tag">Element Plus / Vant UI</span>
          <span class="tech-tag">Go</span>
          <span class="tech-tag">Gin</span>
          <span class="tech-tag">PostgreSQL</span>
        </div>
      </div>

      <div class="info-card">
        <h2>📖 使用说明</h2>
        <ol>
          <li><strong>创建收藏夹</strong> - 点击侧边栏的 "+" 按钮创建新的收藏夹</li>
          <li><strong>选择收藏夹</strong> - 点击收藏夹名称切换当前收藏夹</li>
          <li><strong>创建文件夹</strong> - 在收藏夹内点击 "+" 创建文件夹进行分类</li>
          <li><strong>添加书签</strong> - 点击右上角 "添加书签" 按钮</li>
          <li><strong>搜索书签</strong> - 使用顶部搜索框快速查找书签</li>
          <li><strong>管理书签</strong> - 点击书签卡片上的编辑/删除按钮进行管理</li>
        </ol>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
// 无需导入 AppLayout，由 App.vue 自动提供
</script>

<style scoped>
.about-view {
  max-width: 1000px;
  margin: 0 auto;
  padding: 20px;
}

.about-container h1 {
  font-size: 36px;
  margin-bottom: 40px;
  color: #333;
  text-align: center;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.info-card {
  background: white;
  border-radius: 16px;
  padding: 32px;
  margin-bottom: 28px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  border: 1px solid #f0f0f0;
  transition: all 0.3s ease;
}

.info-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.info-card h2 {
  font-size: 24px;
  margin-bottom: 20px;
  color: #333;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 10px;
}

.info-card p {
  color: #666;
  line-height: 1.8;
  font-size: 15px;
}

.info-card ul,
.info-card ol {
  padding-left: 28px;
  color: #666;
  line-height: 2.2;
  font-size: 15px;
}

.info-card li {
  margin-bottom: 12px;
  transition: all 0.2s ease;
}

.info-card li:hover {
  color: #667eea;
  transform: translateX(4px);
}

.tech-stack {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
}

.tech-tag {
  background: linear-gradient(135deg, #e6f7ff 0%, #bae7ff 100%);
  color: #1890ff;
  padding: 8px 20px;
  border-radius: 20px;
  font-size: 14px;
  font-weight: 500;
  border: 1px solid #91d5ff;
  transition: all 0.3s ease;
  cursor: default;
}

.tech-tag:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(24, 144, 255, 0.3);
  background: linear-gradient(135deg, #bae7ff 0%, #91d5ff 100%);
}

/* 响应式设计 */
@media (min-width: 1920px) {
  .about-view {
    max-width: 1200px;
  }

  .about-container h1 {
    font-size: 42px;
  }

  .info-card {
    padding: 40px;
  }

  .info-card h2 {
    font-size: 28px;
  }
}

@media (max-width: 1440px) {
  .about-view {
    max-width: 900px;
  }
}

@media (max-width: 1024px) {
  .about-view {
    max-width: 100%;
    padding: 16px;
  }

  .about-container h1 {
    font-size: 28px;
  }

  .info-card {
    padding: 24px;
  }
}
</style>
