# Volans

> 飞鱼跃出海面，不是为了逃离海洋，而是为了自由地前行。

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go](https://img.shields.io/badge/Go-1.27-00ADD8?logo=go)](https://go.dev)
[![Vue](https://img.shields.io/badge/Vue-3-42b883?logo=vue.js)](https://vuejs.org)

> 官网：https://hyucn666.github.io/volans/

**Volans 是一个自托管的可视化代理部署管理面板。** 它把「装内核、配节点、开用户、发订阅」这串繁琐的运维动作，压缩成浏览器里点几下鼠标。

几分钟内，一台光秃秃的服务器就能变成一台可以对外发订阅的代理节点。

## 为什么要写它

如果你曾在深夜里，因为 `config.json` 少了一个逗号而对着屏幕反复 `systemctl restart xray`；或者为了给朋友开个账号，SSH 进去手搓 UUID、手动算流量配额——那你大概能理解这个项目为什么存在。

我想做的不是「修风筝」，而是「放风筝」：看清全局，点几下鼠标，剩下的交给它飞。

## 名字

「Volans」是飞鱼座——南天的一群星，得名于跃出水面的飞鱼。飞鱼挣脱海水去滑翔，正如代理让流量摆脱地理的桎梏；航海者借它辨向，也呼应了「为流量指引方向」这件事。

## 特性

- **内核托管**：检测 / 一键安装 / 升级 / 启停 Xray，全程不用碰 SSH 和 JSON
- **多服务器管理**：通过 SSH 集中管理多台节点，远程装内核、下发配置、启停、看流量
- **可视化节点**：VLESS + WS + TLS 图形化配置，改完即热重载，连接不断
- **机场式订阅用户**：一用户一订阅链接，独立流量配额、到期时间、限速，超限自动停
- **并发设备数限制**：限制单个订阅同时在线设备数，超了就温柔地请它下线
- **实时流量**：Xray Stats API + WebSocket，仪表盘实时曲线
- **订阅导出**：V2RayN / Clash / Sing-box 一键生成
- **安全默认**：自动生成持久化密钥、登录失败锁定、登录限流、2FA（TOTP）、多管理员 + RBAC、面板 HTTPS
- **自动证书**：一键从 Let's Encrypt 签发 TLS 证书，告别手动续期
- **告警通知**：用户超限/到期自动停用时，Webhook / Telegram 通知你
- **现代 UI**：Arco Design 极简商务风，暗 / 亮双主题

## 为什么是 Volans

| | Volans | 同类面板 |
|---|---|---|
| 界面 | 现代极简，暗色优先 | 多为传统后台风格 |
| 架构 | 清晰分层，Go 单二进制 | 部分耦合较重 |
| 热重载 | SIGHUP，不中断连接 | 多为整体重启 |
| 多机 | SSH 原生支持 | 多为单机 |
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

默认账号 `admin` / `admin`，仅供开发，生产请第一时间改掉它。

## 部署

一条命令，开箱即用（Ubuntu / Debian）：

```bash
curl -fsSL https://raw.githubusercontent.com/HYuCN666/volans/main/install.sh | bash
```

也可以手动来——从 [Release](https://github.com/HYuCN666/volans/releases) 下载对应平台的包：

```bash
tar -xzf volans-linux-amd64.tar.gz -C /opt
ln -s /opt/volans/volans /usr/local/bin/volans

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

> 安全提示：JWT 密钥自动生成并持久化到 `--data/.secret`；面板 HTTPS 可用 `--tls-cert` / `--tls-key`，建议经 Nginx/Caddy 反代，别裸奔公网。

## 使用流程

1. 登录后在「系统设置」填公网地址 / 域名
2. 「入站节点」新建节点（VLESS + WS + TLS，可一键自动签发证书）
3. 「订阅用户」新建用户，设流量配额和并发设备数
4. 复制订阅链接，发给用户

## 常见问题

**为什么叫飞鱼？** 因为它会飞，又离不开水——像极了我们，既向往自由，又活在现实里。

**安全吗？** 内置密钥持久化、登录锁定/限流、2FA、RBAC。但任何代理服务都请只给信任的人用，并遵守当地法规。

**能管几台机器？** 面板自己跑在一台上，再通过 SSH 管任意多台节点。加机器 = 填个 SSH 地址。

**数据存哪？** 全部在 SQLite（`--data` 目录），迁移就是拷一个文件夹。

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
- [x] ACME 自动证书、告警通知
- [ ] 面板自我更新、旧面板迁移导入

## 许可证

[AGPL-3.0](LICENSE)
