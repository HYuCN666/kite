# Volans

> 飞鱼跃出海面，不是为了逃离海洋，而是为了自由地前行。

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go](https://img.shields.io/badge/Go-1.27-00ADD8?logo=go)](https://go.dev)
[![Vue](https://img.shields.io/badge/Vue-3-42b883?logo=vue.js)](https://vuejs.org)

> 官网：https://hyucn666.github.io/volans/

**Volans** 是一个自托管的可视化代理部署管理面板。它把「装内核、配节点、开用户、发订阅」这串繁琐的运维动作，压缩成浏览器里点几下鼠标。

几分钟内，一台光秃秃的服务器就能变成一台可运营的代理节点。

## 名字

「Volans」是飞鱼座——南天的一群星，得名于跃出水面的飞鱼。飞鱼挣脱海水的束缚去滑翔，正如代理让流量摆脱地理的桎梏；而作为航海者借以辨向的星座，它也呼应我们「为流量指引方向」的使命。

## 特性

- **内核管理**：检测 / 一键安装 / 升级 / 启停 Xray，全程不碰 SSH、不碰 JSON
- **多服务器管理**：通过 SSH 集中管理多台节点，远程安装内核、下发配置、启停
- **可视化节点**：VLESS + WS + TLS 图形化配置，改完即热重载，连接不断
- **机场式订阅用户**：一用户一订阅链接，独立流量配额、到期时间、限速，超限自动停
- **并发设备数限制**：限制单订阅同时在线设备数，超限自动停用
- **实时流量**：基于 Xray Stats API + WebSocket，仪表盘实时曲线
- **订阅导出**：V2RayN / Clash / Sing-box 一键生成
- **安全默认**：自动生成持久化密钥、登录失败锁定、登录限流、面板 HTTPS
- **现代 UI**：Arco Design 极简商务风，暗/亮双主题

## 为什么是 Volans

| | Volans | 同类面板 |
|---|---|---|
| 界面 | 现代极简，暗色优先 | 多为传统后台风格 |
| 架构 | 清晰分层，Go 单二进制 | 部分耦合较重 |
| 热重载 | SIGHUP，不中断连接 | 多为整体重启 |
| 许可 | AGPL-3.0 | 各异 |

## 快速开始（开发）

```bash
# 后端
cd cmd/volans
go run .

# 前端（另开终端）
cd web
npm install
npm run dev
```

默认账号 `admin` / `admin`，仅用于开发，生产请立即修改。

## 部署

详见 [Release](https://github.com/HYuCN666/volans/releases) 的二进制包，或：

```bash
# 下载对应平台包（以 linux-amd64 为例）
tar -xzf volans-linux-amd64.tar.gz -C /opt
ln -s /opt/volans/volans /usr/local/bin/volans

# systemd 服务
cat > /etc/systemd/system/volans.service <<'EOF'
[Unit]
Description=Volans
After=network.target
[Service]
Type=simple
ExecStart=/usr/local/bin/volans --bind 0.0.0.0 --port 8080 --data /etc/volans --web /opt/volans/web/dist
Restart=on-failure
[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload && systemctl enable --now volans
```

> 安全提示：JWT 密钥自动生成并持久化到 `--data/.secret`；面板 HTTPS 用 `--tls-cert` / `--tls-key` 启用，建议经 Nginx/Caddy 反代，而非裸奔公网。

## 使用流程

1. 登录后在「系统设置」填写公网地址/域名
2. 「入站节点」新建节点（VLESS + WS + TLS）
3. 「订阅用户」新建用户，设置流量配额
4. 复制订阅链接，发给用户

## 技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go + Gin + SQLite (WAL) |
| 前端 | Vue 3 + TypeScript + Vite + Arco Design + ECharts |
| 内核 | Xray-core（原生二进制，systemd 托管） |

## 文档

- [产品需求 PRD](docs/PRD.md)
- [架构设计](docs/architecture.md)
- [数据模型](docs/data-model.md)
- [API 设计](docs/api.md)

## Roadmap

- [x] 单机 MVP：内核管理、节点、用户、订阅、实时流量
- [x] 安全加固：密钥持久化、登录锁定/限流、面板 HTTPS
- [x] 并发设备数限制、Sing-box 订阅、多服务器管理（SSH）
- [x] 远程节点流量监控（SSH 隧道采集）
- [x] 多管理员 + RBAC、2FA（TOTP）
- [ ] ACME 自动证书
- [ ] 告警

## 许可证

[AGPL-3.0](LICENSE)
