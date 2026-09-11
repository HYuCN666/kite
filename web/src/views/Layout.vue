<template>
  <a-layout class="layout">
    <a-layout-sider :width="220" class="sider">
      <div class="logo">Kite</div>
      <a-menu :selected-keys="selectedKeys" @menu-item-click="onMenuClick">
        <a-menu-item key="/dashboard">
          <template #icon><icon-dashboard /></template>仪表盘
        </a-menu-item>
        <a-menu-item key="/inbounds">
          <template #icon><icon-share-alt /></template>入站节点
        </a-menu-item>
        <a-menu-item key="/users">
          <template #icon><icon-user-group /></template>订阅用户
        </a-menu-item>
        <a-menu-item key="/settings">
          <template #icon><icon-settings /></template>系统设置
        </a-menu-item>
      </a-menu>
    </a-layout-sider>
    <a-layout>
      <a-layout-header class="header">
        <div class="header-right">
          <a-button shape="circle" size="small" @click="toggleTheme">
            <icon-moon-fill v-if="theme === 'light'" />
            <icon-sun-fill v-else />
          </a-button>
          <a-dropdown>
            <span class="username">{{ auth.username || 'admin' }}</span>
            <template #content>
              <a-doption @click="logout">退出登录</a-doption>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>
      <a-layout-content class="content">
        <router-view />
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  IconDashboard,
  IconShareAlt,
  IconUserGroup,
  IconSettings,
  IconMoonFill,
  IconSunFill,
} from '@arco-design/web-vue/es/icon'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const selectedKeys = computed(() => [route.path])
const theme = ref(localStorage.getItem('kite_theme') || 'dark')

function applyTheme() {
  document.documentElement.setAttribute('arco-theme', theme.value)
  localStorage.setItem('kite_theme', theme.value)
}

function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  applyTheme()
}

function onMenuClick(key: string) {
  router.push(key)
}

function logout() {
  auth.logout()
  router.push('/login')
}

applyTheme()
</script>

<style scoped>
.layout {
  height: 100%;
}

.sider {
  background: var(--kite-surface);
  border-right: 1px solid var(--kite-border);
}

.logo {
  height: 56px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  font-size: 18px;
  font-weight: 700;
  color: var(--kite-primary);
}

.header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: 0 20px;
  background: var(--kite-surface);
  border-bottom: 1px solid var(--kite-border);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.username {
  cursor: pointer;
  color: var(--kite-text);
}

.content {
  overflow: auto;
  background: var(--kite-bg);
}
</style>
