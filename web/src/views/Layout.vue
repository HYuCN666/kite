<template>
  <a-layout class="layout-container">
    <!-- 侧边栏 -->
    <a-layout-sider
      :collapsed="appStore.sidebarCollapsed"
      :collapsible="true"
      :width="220"
      :collapsed-width="64"
      class="layout-sider"
      :hide-trigger="true"
    >
      <!-- Logo 区域 -->
      <div class="logo-wrapper">
        <div class="logo-icon">
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2.2">
            <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <div v-if="!appStore.sidebarCollapsed" class="logo-text">
          <span class="logo-title">Volans</span>
          <span class="logo-tag">Proxy</span>
        </div>
      </div>

      <!-- 导航菜单 -->
      <div class="menu-container">
        <a-menu
          :selected-keys="[activeMenuKey]"
          :auto-open-selected="true"
          @menu-item-click="handleMenuClick"
        >
          <a-menu-item key="Dashboard">
            <template #icon><icon-dashboard /></template>
            仪表盘
          </a-menu-item>
          <a-menu-item key="Inbounds">
            <template #icon><icon-storage /></template>
            入站节点
          </a-menu-item>
          <a-menu-item key="Users">
            <template #icon><icon-user-group /></template>
            订阅用户
          </a-menu-item>
          <a-menu-item key="Settings">
            <template #icon><icon-settings /></template>
            系统设置
          </a-menu-item>
        </a-menu>
      </div>

      <!-- 底部折叠切换 -->
      <div class="sider-footer">
        <a-button type="text" shape="circle" size="small" @click="appStore.toggleSidebar">
          <icon-menu-unfold v-if="appStore.sidebarCollapsed" />
          <icon-menu-fold v-else />
        </a-button>
      </div>
    </a-layout-sider>

    <!-- 主内容区 -->
    <a-layout class="main-layout">
      <!-- 顶部 Header -->
      <a-layout-header class="layout-header">
        <div class="header-left">
          <span class="page-title">{{ currentRouteTitle }}</span>
        </div>

        <div class="header-right">
          <!-- 主题切换按钮 -->
          <a-tooltip :content="appStore.theme === 'dark' ? '切换为浅色模式' : '切换为暗色模式'">
            <a-button
              type="text"
              shape="circle"
              class="header-action-btn"
              @click="appStore.toggleTheme"
            >
              <icon-sun v-if="appStore.theme === 'dark'" />
              <icon-moon v-else />
            </a-button>
          </a-tooltip>

          <a-divider direction="vertical" class="header-divider" />

          <!-- 用户名下拉菜单 -->
          <a-dropdown trigger="hover">
            <div class="user-profile-trigger">
              <a-avatar :size="28" class="user-avatar">
                <icon-user />
              </a-avatar>
              <span class="user-name">{{ userStore.username || 'Admin' }}</span>
              <icon-down class="dropdown-icon" />
            </div>
            <template #content>
              <a-doption @click="router.push('/settings')">
                <template #icon><icon-settings /></template>
                系统设置
              </a-doption>
              <a-doption @click="handleLogout" style="color: rgb(var(--danger-6));">
                <template #icon><icon-export /></template>
                退出登录
              </a-doption>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>

      <!-- 页面内容 -->
      <a-layout-content class="layout-content">
        <div class="content-wrapper">
          <router-view v-slot="{ Component }">
            <transition name="fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </div>
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Modal, Message } from '@arco-design/web-vue'
import { useAppStore } from '@/store/app'
import { useUserStore } from '@/store/user'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()
const userStore = useUserStore()

const activeMenuKey = computed(() => {
  return (route.name as string) || 'Dashboard'
})

const currentRouteTitle = computed(() => {
  return (route.meta?.title as string) || '控制台'
})

const handleMenuClick = (key: string) => {
  router.push({ name: key })
}

const handleLogout = () => {
  Modal.confirm({
    title: '确认退出',
    content: '您确定要退出当前管理会话吗？',
    okText: '退出登录',
    cancelText: '取消',
    okButtonProps: { status: 'danger' },
    onOk: () => {
      userStore.logout()
      Message.success('已安全退出登录')
      router.push('/login')
    }
  })
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
  width: 100vw;
  background-color: var(--bg-app);
}

.layout-sider {
  background-color: var(--bg-card);
  border-right: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  height: 100%;
  transition: width 0.2s cubic-bezier(0.34, 0.69, 0.1, 1);
}

.logo-wrapper {
  height: 60px;
  display: flex;
  align-items: center;
  padding: 0 18px;
  border-bottom: 1px solid var(--border-subtle);
  gap: 12px;
}

.logo-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: rgba(99, 102, 241, 0.12);
  color: var(--primary-6);
  flex-shrink: 0;
}

.logo-text {
  display: flex;
  align-items: center;
  gap: 6px;
  overflow: hidden;
  white-space: nowrap;
}

.logo-title {
  font-size: 16px;
  font-weight: 700;
  letter-spacing: -0.3px;
  color: var(--text-main);
}

.logo-tag {
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 4px;
  background-color: rgba(99, 102, 241, 0.1);
  color: var(--primary-6);
  font-weight: 600;
  text-transform: uppercase;
}

.menu-container {
  flex: 1;
  padding-top: 12px;
  overflow-y: auto;
}

:deep(.arco-menu) {
  background-color: transparent;
}

:deep(.arco-menu-item) {
  margin: 4px 8px;
  border-radius: 6px;
  font-size: 13.5px;
  font-weight: 500;
  height: 38px;
  line-height: 38px;
}

:deep(.arco-menu-selected) {
  background-color: rgba(99, 102, 241, 0.12) !important;
  color: var(--primary-6) !important;
  font-weight: 600;
}

.sider-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--border-subtle);
  display: flex;
  justify-content: flex-end;
}

.main-layout {
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.layout-header {
  height: 60px;
  padding: 0 24px;
  background-color: var(--bg-card);
  border-bottom: 1px solid var(--border-subtle);
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}

.page-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-main);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-action-btn {
  color: var(--text-muted);
  font-size: 16px;
}

.header-action-btn:hover {
  color: var(--text-main);
  background-color: var(--color-fill-2);
}

.header-divider {
  margin: 0 4px;
  border-color: var(--border-subtle);
}

.user-profile-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  border-radius: 6px;
  cursor: pointer;
  user-select: none;
  transition: background-color 0.2s;
}

.user-profile-trigger:hover {
  background-color: var(--color-fill-2);
}

.user-avatar {
  background-color: var(--primary-6);
}

.user-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-main);
}

.dropdown-icon {
  font-size: 11px;
  color: var(--text-muted);
}

.layout-content {
  flex: 1;
  overflow-y: auto;
  background-color: var(--bg-app);
  padding: 24px;
}

.content-wrapper {
  max-width: 1360px;
  margin: 0 auto;
}

/* 路由转场 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
