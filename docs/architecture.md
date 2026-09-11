# Volans 架构设计文档

> 版本：v0.1 · 状态：草稿

## 1. 设计原则

- **进程解耦**：面板进程与 Xray 进程完全独立，面板崩溃不影响转发。
- **单一职责**：分层清晰（transport / api / domain / infra），便于社区贡献。
- **可扩展**：协议、传输层、订阅格式均以接口形式预留扩展点。
- **默认安全**：面板默认仅监听本机，密钥不落明文日志。

## 2. 总体架构

```
┌──────────────────────────────────────────────────────────┐
│                      前端 (web/)                          │
│   Vue 3 + TS + Arco Design · Pinia · ECharts             │
│        REST (JSON)          │          WebSocket          │
└──────────────┬─────────────┴─────────────┬────────────────┘
               │                           │
┌──────────────▼───────────────────────────▼────────────────┐
│                    后端 (Go · Gin)                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │  server  │  │   api    │  │   auth   │  │  stats   │  │
│  │ 路由/WS  │  │ handlers │  │ JWT/锁定 │  │ Stats采集 │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  │
│       │             │             │             │         │
│  ┌────▼─────────────▼─────────────▼─────────────▼────┐    │
│  │                   xray (内核管理)                   │    │
│  │  下载/版本 · 进程控制 · JSON 生成 · 热重载          │    │
│  └────┬──────────────────────────────────┬───────────┘    │
│  ┌────▼───────────┐                 ┌────▼───────────┐    │
│  │  store (SQLite)│                 │  Xray Stats API │    │
│  └────────────────┘                 │  (gRPC, 本机)  │    │
│                                     └────────────────┘    │
└───────────────────────────────────────────────────────────┘
                │                        │
        ┌───────▼───────┐        ┌───────▼────────┐
        │   SQLite 文件  │        │  Xray-core 进程 │
        │  (volans.db)  │        │ (systemd 托管)  │
        └───────────────┘        └────────────────┘
```

## 3. 目录结构

```
volans/
├── cmd/
│   └── volans/             # 入口 main.go（flag 解析、装配）
├── internal/
│   ├── server/            # HTTP 服务、路由注册、WS 升级
│   ├── api/               # REST handlers、请求/响应 DTO
│   │   ├── v1/            # 版本化路由
│   │   └── middleware/    # 鉴权、限流、日志中间件
│   ├── auth/              # 登录、JWT 签发/校验、失败锁定
│   ├── xray/              # 内核：下载、版本、进程、配置生成、热重载
│   │   ├── config/        # Xray JSON 结构体与生成器
│   │   └── proc/          # systemd/进程管理
│   ├── stats/             # Stats API 客户端、聚合、定时采集
│   ├── model/             # 领域模型（struct）
│   ├── store/             # SQLite 数据访问层（DAO）
│   └── config/            # 面板自身配置加载
├── web/                   # Vue 前端
│   ├── src/
│   │   ├── views/         # 页面
│   │   ├── components/    # 通用组件
│   │   ├── api/           # 后端接口封装
│   │   ├── stores/        # Pinia
│   │   └── styles/        # 主题 token
│   └── package.json
├── docs/                  # 本文档目录
├── scripts/               # install.sh / install.ps1
├── Dockerfile
├── docker-compose.yml
├── LICENSE                # AGPL-3.0
└── README.md
```

## 4. 后端分层与职责

| 包 | 职责 |
|---|---|
| `server` | 启动 HTTP/HTTPS 服务，注册 REST 路由与 WS 端点 |
| `api` | 业务 handler，参数校验，调用 domain 层，返回 DTO |
| `auth` | 登录逻辑、bcrypt 密码校验、JWT 签发/校验、失败锁定状态 |
| `xray` | 唯一与 Xray 内核交互的包（下载、进程、JSON 生成、热重载） |
| `stats` | 连接 Xray Stats API，周期性拉取并聚合流量 |
| `model` | 领域结构体，与数据库表对应 |
| `store` | SQLite CRUD、事务、迁移 |
| `config` | 面板配置（端口、数据目录、xray 路径等） |

## 5. 关键流程

### 5.1 内核部署流程

```
用户点击「安装 Xray」
  → xray 包检查本地版本
  → 下载官方 release 二进制（校验校验和）
  → 解压至 /usr/local/bin/xray
  → 生成 systemd unit 文件并 enable
  → 生成初始 config.json（空规则）
  → 启动进程 → 返回状态
```

### 5.2 配置生成与热重载流程

```
用户在 UI 保存入站/用户
  → api 写入 store（事务）
  → api 调用 xray.Rebuild()：
     1. 从 store 读取全部入站/用户
     2. 拼装 Xray config.json（含 stats/policy 配置）
     3. 写入配置文件（权限 0600）
     4. 调用 Xray API（或 signal）触发热重载
  → 返回结果；失败则回滚并提示
```

### 5.3 流量统计流程

```
stats 包启动定时任务（默认 10s）
  → 连接 Xray Stats API（gRPC，127.0.0.1）
  → 拉取 user>>>[email]>>>traffic>>>uplink/downlink
  → 计算增量，更新内存缓存 + 持久化快照
  → 通过 WS 推送增量给前端
  → 检测是否超限 → 超限触发停用（从配置移除用户并热重载）
```

### 5.4 订阅导出流程

```
用户请求 /sub/{token}
  → api 校验令牌 → 查该用户及所属入站
  → 组装 VLESS 分享链接（vless://uuid@host:port?...）
  → 按 Accept/参数输出 V2RayN(base64) 或 Clash(YAML)
```

## 6. 通信协议

- **REST**：JSON 请求/响应，`/api/v1/*`，JWT Bearer 鉴权。
- **WebSocket**：`/ws`，推送三类消息：流量增量、节点状态、日志。

### WebSocket 消息格式

```json
{ "type": "traffic", "data": { "user_id": 1, "uplink": 1024, "downlink": 2048 } }
{ "type": "status", "data": { "running": true, "version": "1.8.x" } }
{ "type": "log", "data": { "level": "info", "message": "..." } }
```

## 7. 安全设计

- 面板默认监听 `127.0.0.1`，公网暴露需显式配置 `--bind 0.0.0.0`。
- Xray 配置文件与证书密钥落盘权限 `0600`。
- 登录失败锁定：同源 IP 连续失败 5 次锁定 15 分钟。
- 订阅令牌为随机高熵字符串，防止被枚举扫描。
- 密码使用 bcrypt 存储。

## 8. 部署方式

1. **单二进制**：`go install` / 下载 release，`volans` 直接运行。
2. **Docker**：镜像内置面板 + 共享 Xray 进程卷。
3. **一键脚本**：`scripts/install.sh`（Ubuntu/Debian）自动下载、建 systemd 服务、开机自启。

## 9. 扩展点

| 扩展点 | 说明 |
|---|---|
| 协议 | `xray/config` 中的 protocol 以接口抽象，新增协议只需实现序列化 |
| 传输层 | 传输层配置结构体独立，可逐步补充 |
| 订阅格式 | 订阅导出器接口化，新增格式实现一个 renderer |
| 存储 | `store` 定义接口，可替换为 MySQL/Postgres |
