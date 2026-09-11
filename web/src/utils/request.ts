import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { Message } from '@arco-design/web-vue'
import type { ApiResponse } from '@/types/api'

// 创建 axios 实例
const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('volans_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const res = response.data
    // 如果返回数据结构不含 code，则直接返回原始数据兼容处理
    if (res && typeof res.code !== 'undefined') {
      if (res.code === 0 || res.code === 200) {
        return res as any
      } else {
        Message.error(res.message || '请求发生错误')
        return Promise.reject(new Error(res.message || 'Error'))
      }
    }
    return res as any
  },
  (error) => {
    if (error.response) {
      const status = error.response.status
      const msg = error.response.data?.message || error.message || '网络请求错误'

      if (status === 401) {
        Message.error('登录状态已失效，请重新登录')
        localStorage.removeItem('volans_token')
        localStorage.removeItem('volans_username')
        if (window.location.pathname !== '/login') {
          window.location.href = '/login'
        }
      } else if (status === 403) {
        Message.error('无权进行该操作')
      } else if (status === 500) {
        Message.error('服务器内部错误: ' + msg)
      } else {
        Message.error(msg)
      }
    } else {
      Message.error('无法连接到服务器，请检查网络或后端服务')
    }
    return Promise.reject(error)
  }
)

export default request
