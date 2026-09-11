import { defineStore } from 'pinia'
import { ref } from 'vue'
import { authApi } from '@/api'
import type { LoginParams } from '@/types/api'

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('volans_token') || '')
  const username = ref<string>(localStorage.getItem('volans_username') || 'Admin')

  const login = async (params: LoginParams) => {
    const res = await authApi.login(params)
    if (res.data?.token) {
      token.value = res.data.token
      username.value = params.username || 'Admin'
      localStorage.setItem('volans_token', res.data.token)
      localStorage.setItem('volans_username', username.value)
    }
    return res
  }

  const logout = () => {
    token.value = ''
    username.value = ''
    localStorage.removeItem('volans_token')
    localStorage.removeItem('volans_username')
  }

  const isAuthenticated = () => {
    return !!token.value
  }

  return {
    token,
    username,
    login,
    logout,
    isAuthenticated
  }
})
