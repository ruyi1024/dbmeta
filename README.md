# DBMeta · Open Source Data Governance Platform

[中文](./README.md) | [English](./README.en.md) | [Changelog](./CHANGELOG.md)

DBMeta 是一个面向数据库治理场景的开源平台，提供从元数据管理、数据质量治理、任务编排到 AI 辅助分析的统一能力。

本仓库为 **核心开源仓库**，包含后端服务、前端工程与部署资产，适用于企业内部自托管与二次开发。

[![GitHub stars](https://img.shields.io/github/stars/ruyi1024/dbmeta?style=social)](https://github.com/ruyi1024/dbmeta/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/ruyi1024/dbmeta?style=social)](https://github.com/ruyi1024/dbmeta/network/members)
[![Website](https://img.shields.io/badge/Website-dbmeta.com-blue?style=flat&logo=googlechrome&logoColor=white)](https://www.dbmeta.com)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev/)
[![Last commit](https://img.shields.io/github/last-commit/ruyi1024/dbmeta)](https://github.com/ruyi1024/dbmeta/commits/main)
[![Latest release](https://img.shields.io/github/v/release/ruyi1024/dbmeta?style=flat)](https://github.com/ruyi1024/dbmeta/releases/latest)


---

## 项目概览

- **治理资产统一可见**：覆盖数据源、实例、库、表、字段与业务信息
- **治理流程闭环可运营**：规则、任务、问题、看板完整串联
- **AI 能力深度融合**：会话、规则、模型管理与路由策略一体化
- **全栈工程化交付**：后端 + 前端 + Docker，支持快速落地

---

## 核心能力

| 模块 | 能力 |
|---|---|
| 元数据治理 | 数据源管理、库表字段管理、业务信息维护 |
| 数据查询 | 查询入口、收藏、权限边界控制 |
| 数据质量 | 规则、任务、问题、仪表盘 |
| AI 能力 | AI 对话、规则、模型配置、会话管理 |
| 容量分析 | 容量统计、增长分析、Top-N 视图 |
| 系统任务 | 定时执行、任务日志、任务配置 |

---

## Quick Start

### 环境要求

- Go 1.19+
- MySQL 8+
- Redis 6+
- Node.js / pnpm（前端开发时需要）

### 本地运行（后端）

```bash
go mod tidy
go run . -c ./setting.yml
```

> 建议先复制 `setting.example.yml` 生成本地配置，再填写数据库连接信息。

### Docker 一键部署

```bash
docker compose -f docker/docker-compose.yml up -d --build
```

默认访问：`http://127.0.0.1:8086`

---

## 项目结构

```text
dbmeta-core/
├─ app/                 # 启动与引导
├─ router/              # 路由注册
├─ setting/             # 配置解析
├─ src/
│  ├─ controller/       # HTTP 控制器
│  ├─ service/          # 业务服务层
│  ├─ model/            # 数据模型
│  ├─ database/         # 数据库初始化与迁移
│  ├─ task/             # 定时与后台任务
│  ├─ module/           # 模块注册与扩展点
│  └─ ...
├─ frontend/            # 前端工程（Vben Monorepo）
├─ webassets/           # 嵌入式静态资源
└─ docker/              # Docker 部署文件
```

---

## Tech Stack

| 类别 | 技术栈 |
|---|---|
| 后端 | Go、Gin、GORM、定时任务系统 |
| 前端 | Vue / Vben Admin（Monorepo） |
| 数据层 | MySQL、Redis |
| AI 能力 | 多模型接入、路由与会话管理 |
| 部署 | Docker Compose、嵌入式静态资源发布 |

---

## 配置说明

- 示例配置：`setting.example.yml`
- 本地配置建议使用：`setting.yml`（已在 `.gitignore` 中）
- 业务通知相关配置已迁移到数据库配置表维护
- 生产环境请使用独立配置文件与密钥管理策略

---

## 开发指南

### 后端开发

- 入口文件：`main.go`
- 核心引导：`app/bootstrap.go`
- 路由定义：`router/router.go`

### 前端开发

- 前端位于 `frontend/`
- 参考前端子工程内文档与脚本启动开发环境

### 质量检查

```bash
go build .
```

---

## FAQ

### 启动端口冲突

在配置中设置 `server.addr` 为其他端口后重启。

### 前端接口指向错误

检查前端开发环境代理配置，确认目标后端地址正确。

### 启动后页面空白

确认 `webassets` 静态资源与当前后端版本匹配。

---

## 贡献

欢迎通过 Issue / Pull Request 参与改进：

- Bug 修复
- 文档完善
- 治理规则与任务扩展
- AI 能力优化与模型适配

建议提交前执行基础构建校验并附带测试说明。

---

## License

AGPL-3.0 仅限非商业用途。任何商业用途均需获得商业授权。

| 使用场景 | 是否允许 |
|---|---|
| 个人 / 研究 / 教育 | 是 |
| 自托管（非商业） | 是，需保留署名 |
| Fork 并修改（非商业） | 是，需按 AGPL-3.0 开源源代码 |
| 商业用途 / SaaS / 品牌重塑 | 需要商业授权 |

完整条款请参见 `LICENSE`。如需商业授权，请联系维护者。

Copyright (C) 2026 DBMETA.COM All rights reserved.

---

## Author

**DBMeta Team** — [DBMETA](https://www.dbmeta.com)

---

## Star History

<a href="https://www.star-history.com/?type=date&repos=ruyi1024%2Fdbmeta">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=ruyi1024/dbmeta&type=date&theme=dark&legend=top-left" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=ruyi1024/dbmeta&type=date&legend=top-left" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=ruyi1024/dbmeta&type=date&legend=top-left" />
 </picture>
</a>