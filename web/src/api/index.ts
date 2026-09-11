import request from '@/utils/request'
import type {
  ApiResponse,
  LoginParams,
  LoginResult,
  UpdatePasswordParams,
  NodeStatus,
  InboundItem,
  UserItem,
  ServerItem,
  SubscriptionResult,
  StatsOverview,
  SystemSettings
} from '@/types/api'

// 1. 认证与账户
export const authApi = {
  login: (data: LoginParams) =>
    request.post<any, ApiResponse<LoginResult>>('/api/v1/auth/login', data),

  updatePassword: (data: UpdatePasswordParams) =>
    request.put<any, ApiResponse<null>>('/api/v1/auth/password', data)
}

// 2. 节点与状态
export const nodeApi = {
  getStatus: () =>
    request.get<any, ApiResponse<NodeStatus>>('/api/v1/node/status'),

  getStatsOverview: () =>
    request.get<any, ApiResponse<StatsOverview>>('/api/v1/stats/overview')
}

// 3. 入站节点管理
export const inboundApi = {
  getList: () =>
    request.get<any, ApiResponse<InboundItem[]>>('/api/v1/inbounds'),

  getById: (id: number | string) =>
    request.get<any, ApiResponse<InboundItem>>(`/api/v1/inbounds/${id}`),

  create: (data: Partial<InboundItem>) =>
    request.post<any, ApiResponse<InboundItem>>('/api/v1/inbounds', data),

  update: (id: number | string, data: Partial<InboundItem>) =>
    request.put<any, ApiResponse<InboundItem>>(`/api/v1/inbounds/${id}`, data),

  delete: (id: number | string) =>
    request.delete<any, ApiResponse<null>>(`/api/v1/inbounds/${id}`),

  updateStatus: (id: number | string, enabled: boolean) =>
    request.patch<any, ApiResponse<null>>(`/api/v1/inbounds/${id}/status`, { enabled })
}

// 4. 用户与订阅管理
export const userApi = {
  getList: (params?: { inbound_id?: number | string }) =>
    request.get<any, ApiResponse<UserItem[]>>('/api/v1/users', { params }),

  getById: (id: number | string) =>
    request.get<any, ApiResponse<UserItem>>(`/api/v1/users/${id}`),

  create: (data: Partial<UserItem>) =>
    request.post<any, ApiResponse<UserItem>>('/api/v1/users', data),

  update: (id: number | string, data: Partial<UserItem>) =>
    request.put<any, ApiResponse<UserItem>>(`/api/v1/users/${id}`, data),

  delete: (id: number | string) =>
    request.delete<any, ApiResponse<null>>(`/api/v1/users/${id}`),

  updateStatus: (id: number | string, enabled: boolean) =>
    request.patch<any, ApiResponse<null>>(`/api/v1/users/${id}/status`, { enabled }),

  resetTraffic: (id: number | string) =>
    request.post<any, ApiResponse<null>>(`/api/v1/users/${id}/reset-traffic`),

  getSubscription: (id: number | string) =>
    request.get<any, ApiResponse<SubscriptionResult>>(`/api/v1/users/${id}/subscription`)
}

// 5. 系统设置
export const settingsApi = {
  getSettings: () =>
    request.get<any, ApiResponse<SystemSettings>>('/api/v1/settings'),

  updateSettings: (data: SystemSettings) =>
    request.put<any, ApiResponse<SystemSettings>>('/api/v1/settings', data)
}

// 6. 节点服务器管理
export const serverApi = {
  getList: () =>
    request.get<any, ApiResponse<ServerItem[]>>('/api/v1/servers'),

  create: (data: any) =>
    request.post<any, ApiResponse<{ id: number }>>('/api/v1/servers', data),

  update: (id: number | string, data: any) =>
    request.put<any, ApiResponse<null>>(`/api/v1/servers/${id}`, data),

  delete: (id: number | string) =>
    request.delete<any, ApiResponse<null>>(`/api/v1/servers/${id}`),

  test: (data: any) =>
    request.post<any, ApiResponse<{ connected: boolean }>>('/api/v1/servers/test', data),

  getStatus: (id: number | string) =>
    request.get<any, ApiResponse<{ installed: boolean; version: string; running: boolean }>>(`/api/v1/servers/${id}/status`),

  install: (id: number | string) =>
    request.post<any, ApiResponse<null>>(`/api/v1/servers/${id}/install`),

  control: (id: number | string, action: string) =>
    request.post<any, ApiResponse<null>>(`/api/v1/servers/${id}/control`, { action })
}
