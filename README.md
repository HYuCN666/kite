# Kite

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

一个自托管的**可视化代理部署管理面板**，用于部署、配置并运营 [Xray-core](https://github.com/XTLS/Xray-core) 正向代理服务。提供现代化的 Web 界面，无需手工编辑 JSON 配置。

> 项目处于早期开发阶段（MVP）。API 与数据库结构可能变动。

## 特性

- 图形化管理 Xray 内核（检测 / 安装 / 升级 / 启停）
- 入站节点可视化配置（VLESS + WS + TLS）
- 机场式订阅用户：一用户一链接、流量配额、到期、限速
- 基于 Xray Stats API 的实时流量统计
- V2RayN / Clash 订阅导出
- 现代 Web UI（Arco Design，极简商务风，暗/亮主题）

## 技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go + Gin + SQLite |
| 前端 | Vue 3 + TypeScript + Vite + Arco Design + Pinia + ECharts |
| 内核 | Xray-core（原生二进制） |

## 文档

- [产品需求 PRD](docs/PRD.md)
- [架构设计](docs/architecture.md)
- [数据模型](docs/data-model.md)
- [API 设计](docs/api.md)

## 开发

### 后端

```bash
cd cmd/kite
go run .
```

### 前端

```bash
cd web
npm install
npm run dev
```

## 部署（Ubuntu/Debian）

Kite 通过 systemd 管理 Xray 内核，建议以 root 权限运行。

### 方式一：下载 Release 二进制

1. 在 [Releases](https://github.com/HYuCN666/kite/releases) 下载对应平台的 `kite-linux-*.tar.gz`。
2. 解压并放置：

```bash
tar -xzf kite-linux-amd64.tar.gz -C /opt
ln -s /opt/kite/kite /usr/local/bin/kite
```

3. 创建 systemd 服务：

```bash
cat > /etc/systemd/system/kite.service <<'EOF'
[Unit]
Description=Kite
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/kite --bind 0.0.0.0 --port 8080 --data /etc/kite --web /opt/kite/web/dist
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now kite
```

4. 浏览器访问 `http://服务器IP:8080`，默认账号 `admin` / `admin`（首次登录后请修改密码）。

> **安全提示**：JWT 密钥会自动生成并持久化到 `--data` 目录下的 `.secret` 文件（无需手动配置）；也可通过 `KITE_SECRET` 环境变量指定。面板 HTTPS 通过 `--tls-cert` / `--tls-key` 启用，建议经 Nginx/Caddy 反代添加 TLS，而非直接暴露公网。

### 方式二：Docker

```bash
docker compose up -d
```

> 注意：Docker 容器内无法直接调用宿主机 systemd 管理 Xray 进程。容器方式下请将 Xray 单独部署在宿主机，或改用原生部署。

## 使用流程

1. 登录后在「系统设置」填写公网地址/域名。
2. 「入站节点」→ 新建节点（VLESS + WS + TLS）。
3. 「订阅用户」→ 新建用户，设置流量配额。
4. 点击「复制订阅」发给用户，或直接访问订阅链接。

## 许可证

[AGPL-3.0](LICENSE)
