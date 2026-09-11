import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('volans_token') || '',
    username: localStorage.getItem('volans_username') || '',
  }),
  actions: {
    setToken(token: string) {
      this.token = token
      localStorage.setItem('volans_token', token)
    },
    setUsername(username: string) {
      this.username = username
      localStorage.setItem('volans_username', username)
    },
    logout() {
      this.token = ''
      this.username = ''
      localStorage.removeItem('volans_token')
      localStorage.removeItem('volans_username')
    },
  },
})
