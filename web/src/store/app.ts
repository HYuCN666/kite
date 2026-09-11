import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ThemeMode = 'dark' | 'light'

export const useAppStore = defineStore('app', () => {
  // 默认暗色主题
  const savedTheme = (localStorage.getItem('volans_theme') as ThemeMode) || 'dark'
  const theme = ref<ThemeMode>(savedTheme)
  const sidebarCollapsed = ref<boolean>(false)

  // 应用主题到 html 标签
  const applyTheme = (mode: ThemeMode) => {
    theme.value = mode
    localStorage.setItem('volans_theme', mode)
    if (mode === 'dark') {
      document.documentElement.setAttribute('arco-theme', 'dark')
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.removeAttribute('arco-theme')
      document.documentElement.classList.remove('dark')
    }
  }

  // 切换主题
  const toggleTheme = () => {
    const next = theme.value === 'dark' ? 'light' : 'dark'
    applyTheme(next)
  }

  // 折叠侧边栏
  const toggleSidebar = () => {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  // 初始化主题
  applyTheme(theme.value)

  return {
    theme,
    sidebarCollapsed,
    applyTheme,
    toggleTheme,
    toggleSidebar
  }
})
