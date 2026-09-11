import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('kite_token') || '',
    username: localStorage.getItem('kite_username') || '',
  }),
  actions: {
    setToken(token: string) {
      this.token = token
      localStorage.setItem('kite_token', token)
    },
    setUsername(username: string) {
      this.username = username
      localStorage.setItem('kite_username', username)
    },
    logout() {
      this.token = ''
      this.username = ''
      localStorage.removeItem('kite_token')
      localStorage.removeItem('kite_username')
    },
  },
})
