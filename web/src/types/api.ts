// API 统一响应结构
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 登录参数与结果
export interface LoginParams {
  username: string
  password: string
  code?: string
}

export interface LoginResult {
  token: string
}

// 修改密码
export interface UpdatePasswordParams {
  old_password: string
  new_password: string
}

// 节点与系统运行状态
export interface SystemMetrics {
  cpu: number // 百分比
  mem: number // 百分比
  disk: number // 百分比
  load: number | string // 系统负载
}

export interface NodeStatus {
  installed: boolean
  version: string
  running: boolean
  system: SystemMetrics
}

// 入站节点流设置
export interface StreamSettings {
  path?: string
}

// 入站节点
export interface InboundItem {
  id?: number | string
  server_id?: number | string
  remark: string
  protocol: 'vless' | string
  port: number
  listen: string
  transport: 'tcp' | 'ws'
  stream_settings?: StreamSettings
  tls_enabled: boolean
  tls_cert?: string
  tls_key?: string
  tls_server_name?: string
  enable_sniffing: boolean
  enabled: boolean
  created_at?: string
  updated_at?: string
}

// 订阅用户
export interface UserItem {
  id?: number | string
  inbound_id: number | string
  remark: string
  quota_bytes: number // 总配额（字节）
  used_uplink: number // 上行用量（字节）
  used_downlink: number // 下行用量（字节）
  speed_limit_uplink: number // 限速（B/s 或 Mbps）
  speed_limit_downlink: number
  max_devices: number // 并发设备数限制（0 为不限）
  expire_at: string | number // 到期时间
  enabled: boolean
  created_at?: string
}

// 节点服务器
export interface ServerItem {
  id: number | string
  name: string
  host: string
  port?: number
  username?: string
  auth_type?: string
  enabled: boolean
}

// 订阅链接响应
export interface SubscriptionResult {
  token: string
  url: string
}

// 流量统计概览
export interface StatsOverview {
  total_uplink: number
  total_downlink: number
  active_users: number
}

// 系统设置
export interface SystemSettings {
  public_host?: string
  [key: string]: any
}

// 管理员
export interface AdminItem {
  id: number | string
  username: string
  role: string
  totp_enabled: boolean
  created_at?: string
}

// WebSocket 实时流量数据
export interface WsTrafficData {
  uplink: number // 实时上行速率 bps/Bps
  downlink: number // 实时下行速率 bps/Bps
  time: string | number
}

export interface WsMessage {
  type: 'traffic' | string
  data: WsTrafficData
}
