# Volans 数据模型设计

> 版本：v0.1 · 状态：草稿 · 数据库：SQLite（默认）

## 1. 约定

- 所有表包含 `id` 自增主键、`created_at`、`updated_at` 时间戳。
- 金额/流量单位统一为**字节（int64）**，前端负责展示换算。
- 时间统一 UTC 存储，展示层转换时区。
- 布尔用 `INTEGER (0/1)`。

## 2. ER 关系概览

```
settings (面板配置, 单行)
admins (管理员, 单管理员 MVP)
inbounds ──1:N── users (用户归属入站)
users ──1:1── subscription (订阅令牌)
traffic_snapshots (流量快照, 按 user 聚合)
```

## 3. 表结构

### 3.1 `settings` — 面板配置

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | INTEGER | PK | |
| key | TEXT | UNIQUE NOT NULL | 配置键 |
| value | TEXT | | 配置值（JSON 序列化） |

常用键：`panel_port`、`panel_bind`、`xray_path`、`xray_version`、`cert_path`、`key_path`、`theme` 等。

### 3.2 `admins` — 管理员

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | INTEGER | PK | |
| username | TEXT | UNIQUE NOT NULL | 登录名 |
| password_hash | TEXT | NOT NULL | bcrypt |
| last_login_at | DATETIME | | 最近登录 |
| failed_attempts | INTEGER | DEFAULT 0 | 连续失败次数 |
| locked_until | DATETIME | NULL | 锁定截止时间 |

### 3.3 `inbounds` — 入站节点

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | INTEGER | PK | |
| tag | TEXT | NOT NULL | Xray 入站标识（如 `inbound-1`） |
| remark | TEXT | | 备注/显示名 |
| protocol | TEXT | NOT NULL | `vless` / `vmess` / ... |
| port | INTEGER | NOT NULL | 监听端口 |
| listen | TEXT | DEFAULT `0.0.0.0` | 监听地址 |
| transport | TEXT | NOT NULL | `ws` / `tcp` / `grpc` ... |
| stream_settings | TEXT | | 传输层参数（JSON） |
| tls_enabled | INTEGER | DEFAULT 0 | 是否启用 TLS |
| tls_cert | TEXT | | 证书路径 |
| tls_key | TEXT | | 私钥路径 |
| tls_server_name | TEXT | | SNI |
| enable_sniffing | INTEGER | DEFAULT 1 | |
| enabled | INTEGER | DEFAULT 1 | 是否启用 |
| created_at / updated_at | DATETIME | | |

> `stream_settings` 示例（WS）：
> `{"network":"ws","wsSettings":{"path":"/ws","headers":{}}}`

### 3.4 `users` — 订阅用户

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | INTEGER | PK | |
| inbound_id | INTEGER | FK → inbounds.id | 归属入站 |
| email | TEXT | UNIQUE NOT NULL | Xray 用户标识（用于 stats 关联） |
| uuid | TEXT | UNIQUE NOT NULL | 连接凭证 UUID |
| remark | TEXT | | 备注 |
| quota_bytes | INTEGER | NOT NULL | 流量配额（0 = 不限） |
| used_uplink | INTEGER | DEFAULT 0 | 已用上行 |
| used_downlink | INTEGER | DEFAULT 0 | 已用下行 |
| speed_limit_uplink | INTEGER | DEFAULT 0 | 上行限速（B/s，0=不限） |
| speed_limit_downlink | INTEGER | DEFAULT 0 | 下行限速 |
| expire_at | DATETIME | NULL | 到期时间 |
| enabled | INTEGER | DEFAULT 1 | 是否启用 |
| created_at / updated_at | DATETIME | | |

> `used_uplink + used_downlink` 由 stats 采集持续更新；超限时置 `enabled=0` 并触发配置重载。

### 3.5 `subscriptions` — 订阅令牌

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | INTEGER | PK | |
| user_id | INTEGER | FK → users.id, UNIQUE | 一用户一链接 |
| token | TEXT | UNIQUE NOT NULL | 随机高熵令牌 |
| created_at / updated_at | DATETIME | | |

### 3.6 `traffic_snapshots` — 流量快照

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | INTEGER | PK | |
| user_id | INTEGER | FK → users.id | |
| inbound_id | INTEGER | FK → inbounds.id | |
| uplink | INTEGER | | 该周期上行增量 |
| downlink | INTEGER | | 该周期下行增量 |
| recorded_at | DATETIME | INDEX | 采集时间 |

> 用于历史流量图表；MVP 可只保留内存增量 + 聚合表，快照表可按需裁剪（定期清理）。

## 4. 索引

| 表 | 索引 | 用途 |
|---|---|---|
| users | `inbound_id` | 按入站查用户 |
| users | `email` (unique) | stats 关联 |
| subscriptions | `token` (unique) | 订阅快速校验 |
| traffic_snapshots | `(user_id, recorded_at)` | 时间范围查询 |

## 5. 枚举约定

| 枚举 | 取值 |
|---|---|
| protocol | `vless`、`vmess`（预留 `trojan`、`shadowsocks`） |
| transport | `tcp`、`ws`、`grpc`、`h2`、`quic`、`mkcp` |
| 订阅格式 | `v2rayn`、`clash`（预留 `singbox`） |

## 6. 迁移策略

- 使用嵌入式迁移（如 `golang-migrate` 或自定义 `schema_migrations` 表），启动时自动升级。
- 版本号递增，`docs/` 同步更新本文档。
