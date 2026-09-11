# Volans API 设计文档

> 版本：v0.1 · 状态：草稿

## 1. 约定

- Base：`/api/v1`
- 格式：请求/响应均 `application/json`
- 鉴权：`Authorization: Bearer <JWT>`（除登录与订阅端点外）
- 统一响应包裹：

```json
{ "code": 0, "message": "ok", "data": { } }
```

- 错误：`code` 非 0，`message` 为人可读错误信息。

### 错误码

| code | 含义 |
|---|---|
| 0 | 成功 |
| 40001 | 参数错误 |
| 40100 | 未认证 / token 失效 |
| 40300 | 无权限 / 账号锁定 |
| 40400 | 资源不存在 |
| 40900 | 冲突（端口占用等） |
| 50000 | 服务器内部错误 |
| 50010 | Xray 操作失败 |

## 2. 认证

### 2.1 登录

```
POST /api/v1/auth/login
```

请求：
```json
{ "username": "admin", "password": "secret" }
```

响应：
```json
{ "code": 0, "data": { "token": "<jwt>", "expires_at": "..." } }
```

### 2.2 修改密码

```
PUT /api/v1/auth/password
```

请求：`{ "old_password": "...", "new_password": "..." }`

### 2.3 当前用户信息

```
GET /api/v1/auth/me
```

## 3. 节点 / 内核管理

### 3.1 获取内核状态

```
GET /api/v1/node/status
```

响应 `data`：
```json
{
  "installed": true,
  "version": "1.8.23",
  "running": true,
  "latest_version": "1.8.24",
  "system": { "cpu": 12.5, "mem": 40.2, "disk": 31.0, "load": 0.4 }
}
```

### 3.2 安装 / 升级内核

```
POST /api/v1/node/install
```

请求：`{ "version": "latest" }`

### 3.3 进程控制

```
POST /api/v1/node/control
```

请求：`{ "action": "start" | "stop" | "restart" }`

### 3.4 获取日志

```
GET /api/v1/node/logs?lines=200
```

## 4. 入站管理

### 4.1 列表

```
GET /api/v1/inbounds
```

响应 `data`：
```json
[
  {
    "id": 1, "remark": "香港节点", "protocol": "vless",
    "port": 443, "transport": "ws", "tls_enabled": true,
    "enabled": true, "user_count": 3
  }
]
```

### 4.2 详情

```
GET /api/v1/inbounds/{id}
```

### 4.3 创建

```
POST /api/v1/inbounds
```

请求：
```json
{
  "remark": "香港节点", "protocol": "vless", "port": 443,
  "listen": "0.0.0.0", "transport": "ws",
  "stream_settings": { "path": "/ws" },
  "tls_enabled": true, "tls_cert": "/etc/xray/cert.pem",
  "tls_key": "/etc/xray/key.pem", "tls_server_name": "example.com"
}
```

### 4.4 更新

```
PUT /api/v1/inbounds/{id}
```

### 4.5 删除 / 启停

```
DELETE /api/v1/inbounds/{id}
PATCH /api/v1/inbounds/{id}/status   { "enabled": false }
```

## 5. 用户管理

### 5.1 列表

```
GET /api/v1/users?inbound_id={id}
```

响应 `data`：
```json
[
  {
    "id": 1, "inbound_id": 1, "remark": "客户A", "uuid": "...",
    "quota_bytes": 107374182400, "used_bytes": 123456,
    "speed_limit": 0, "expire_at": "2026-12-31T00:00:00Z",
    "enabled": true, "subscription_token": "abc..."
  }
]
```

### 5.2 创建

```
POST /api/v1/users
```

请求：
```json
{
  "inbound_id": 1, "remark": "客户A",
  "quota_bytes": 107374182400,
  "speed_limit_uplink": 0, "speed_limit_downlink": 0,
  "expire_at": "2026-12-31T00:00:00Z"
}
```

> `uuid` 由后端生成；创建时自动生成订阅令牌。

### 5.3 更新 / 删除 / 启停

```
PUT /api/v1/users/{id}
DELETE /api/v1/users/{id}
PATCH /api/v1/users/{id}/status   { "enabled": false }
```

### 5.4 流量重置

```
POST /api/v1/users/{id}/reset-traffic
```

## 6. 订阅

### 6.1 获取订阅（公开，无需鉴权）

```
GET /sub/{token}?format=v2rayn|clash
```

- `format` 缺省 `v2rayn`。
- 响应为订阅内容，Content-Type 按格式区分。

### 6.2 获取订阅链接（面板内展示）

```
GET /api/v1/users/{id}/subscription
```

响应 `data`：
```json
{ "token": "abc...", "url": "https://panel.example.com/sub/abc...", "qrcode": "..." }
```

## 7. 流量统计

### 7.1 汇总统计

```
GET /api/v1/stats/overview
```

响应 `data`：
```json
{
  "total_uplink": 1000, "total_downlink": 2000,
  "active_users": 3, "active_connections": 5
}
```

### 7.2 用户流量历史

```
GET /api/v1/stats/users/{id}/traffic?from=&to=&interval=hour
```

### 7.3 实时推送

```
WS /ws?token=<jwt>
```

推送消息见 `architecture.md` 第 6 节。

## 8. 系统设置

```
GET /api/v1/settings
PUT /api/v1/settings
```

请求示例：
```json
{ "panel_port": 8080, "panel_bind": "0.0.0.0", "theme": "dark" }
```

## 9. OpenAPI

- 计划使用 `swaggo/swag` 生成 Swagger 文档，访问 `/swagger/*`。
- 本文档为设计基准，实现时以生成的 OpenAPI 为准。
