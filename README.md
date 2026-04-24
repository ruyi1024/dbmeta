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

[查看完整界面快照](./SCREENIMG.md)

![DBMeta 首页预览](./docs/images/image1.png)
---

## 项目概览

- **治理资产统一可见**：覆盖数据源、实例、库、表、字段与业务信息，形成可检索、可追踪的统一元数据视图
- **数据统一查询与分析**：提供统一查询入口、收藏与权限边界控制，支持跨库快速检索与结果分析
- **数据质量与巡检闭环**：支持质量规则、质量任务、问题跟踪、巡检报告与治理看板，形成可运营的质量治理体系
- **容量管理与趋势洞察**：支持库表容量统计、增长趋势、Top-N 分析与容量巡检，辅助容量规划与治理决策
- **自动化调度与任务编排**：内置周期任务与手动执行能力，支持任务日志、状态追踪与可扩展任务模型
- **AI 能力深度融合**：支持 AI 对话、规则、模型管理与路由策略，兼容本地模型接入，降低数据治理与分析门槛

---

## 支持的数据源

当前版本支持以下数据源类型：

- MySQL
- MariaDB
- GreatSQL
- TiDB
- Doris
- OceanBase
- ClickHouse
- Oracle
- PostgreSQL
- SQLServer
- MongoDB
- Redis

---

## 核心能力

| 模块 | 能力 |
|---|---|
| 工作台 | 首页总览、关键指标、常用入口 |
| 元数据管理 | 实例、库、表、字段、业务信息、库与业务关系、元数据质量视图 |
| 数据统一查询 | 查询入口、收藏、权限边界控制、跨数据源检索 |
| 容量管理 | 容量总览、库容量查询、表容量查询、增长趋势与 Top-N 分析 |
| 数据质量 | 质量总览、问题列表、规则管理、任务执行与追踪 |
| 数据巡检 | 巡检报告、巡检配置、结果跟踪 |
| 自动化任务 | 任务列表、任务参数、执行日志、调度运行状态 |
| 配置中心 | IDC、环境、数据源、通知、AI 模型配置 |
| 用户与权限 | 用户管理、登录会话与访问控制 |
| AI 能力 | AI 对话、规则管理、会话管理、模型路由与本地模型接入 |
| 扩展模块（按版本） | 数据安全（权限/敏感数据）、审计、洞察分析等企业能力 |

---

## 安装与部署

### 0) 环境要求

- Go `1.24+`（`go.mod` 当前为 `go 1.24.1`）
- MySQL `8+`
- Redis `6+`
- Node.js `20.19+ / 22.18+ / 24+`
- pnpm `>=10`（仓库使用 `pnpm@10.33.0`）
- Docker / Docker Compose（使用容器部署时）

### 1) 后端源码部署运行（本地开发）

1. 克隆并进入项目目录
2. 复制配置文件并按实际环境修改
3. 安装依赖并启动

```bash
git clone https://github.com/ruyi1024/dbmeta.git
cd dbmeta
cp setting.example.yml setting.yml
go mod tidy
go run . -c ./setting.yml
```

关键配置项（`setting.yml`）：
- `dataSource.host/port/user/password/database`：MySQL 连接
- `dataSource.redisHost/redisPort/redisPassword`：Redis 连接
- `server.addr`：服务端口（不配时开源版默认 `:8086`）

启动后访问：`http://127.0.0.1:8086`

### 2) 前端编译并集成到后端（webassets）

> 本仓库采用“后端内嵌前端静态资源”模式，编译后需同步到 `webassets/`。

1. 编译前端（web-antd）

```bash
cd frontend
pnpm install
pnpm run build:antd
```

2. 将前端产物同步到后端静态目录（在仓库根目录执行）

```bash
rm -rf webassets/static
mkdir -p webassets/static
cp -a frontend/apps/web-antd/dist/. webassets/static/
mv webassets/static/index.html webassets/index.html
```

3. 启动后端（会加载 `webassets`）

```bash
go run . -c ./setting.yml
```

Windows PowerShell 可参考：

```powershell
Remove-Item -Recurse -Force webassets\static
New-Item -ItemType Directory -Path webassets\static | Out-Null
Copy-Item -Recurse -Force frontend\apps\web-antd\dist\* webassets\static\
Move-Item -Force webassets\static\index.html webassets\index.html
```

### 3) Docker 一键部署（推荐体验）

在仓库根目录执行：

```bash
docker compose -f docker/docker-compose.yml up -d --build
```

该命令会自动完成：
- 启动 MySQL 8 与 Redis 7
- 构建前端并打包到 `webassets`
- 编译并启动后端容器

常用排查命令：

```bash
docker compose -f docker/docker-compose.yml ps
docker compose -f docker/docker-compose.yml logs -f app
docker compose -f docker/docker-compose.yml down
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

### 前端产物管理（B 方案）

- `webassets/` 作为前端构建产物目录，默认保留在仓库中，保障后端开箱即用。
- 允许提交：`webassets/index.html` 与 `webassets/static/**`。
- 非前端发布类改动（如纯后端、文档、脚本）不应携带 `webassets` 变更。
- 若 PR 包含 `webassets`，请在 PR 描述中说明对应前端源码改动范围与构建命令。
- 建议仅在发版、前端功能上线或前端依赖升级时同步更新 `webassets`。

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