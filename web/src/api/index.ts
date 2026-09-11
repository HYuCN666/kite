import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

client.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`
  }
  return config
})

client.interceptors.response.use(
  (res) => res.data,
  (err) => {
    if (err.response?.status === 401) {
      const auth = useAuthStore()
      auth.logout()
      window.location.href = '/login'
    }
    return Promise.reject(err)
  },
)

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

export function get<T = unknown>(url: string, config?: any): Promise<ApiResponse<T>> {
  return client.get(url, config) as unknown as Promise<ApiResponse<T>>
}

export function post<T = unknown>(url: string, data?: any, config?: any): Promise<ApiResponse<T>> {
  return client.post(url, data, config) as unknown as Promise<ApiResponse<T>>
}

export function put<T = unknown>(url: string, data?: any, config?: any): Promise<ApiResponse<T>> {
  return client.put(url, data, config) as unknown as Promise<ApiResponse<T>>
}

export function patch<T = unknown>(url: string, data?: any, config?: any): Promise<ApiResponse<T>> {
  return client.patch(url, data, config) as unknown as Promise<ApiResponse<T>>
}

export function del<T = unknown>(url: string, config?: any): Promise<ApiResponse<T>> {
  return client.delete(url, config) as unknown as Promise<ApiResponse<T>>
}

export default client
