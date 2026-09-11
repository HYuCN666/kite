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
cd cmd/xpanel
go run .
```

### 前端

```bash
cd web
npm install
npm run dev
```

## 许可证

[AGPL-3.0](LICENSE)
