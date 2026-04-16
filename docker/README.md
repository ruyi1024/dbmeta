# 开源版 Docker 一键部署

镜像内包含：**Ant Design 前端（嵌入 `/public/`）** + **Go API（`-tags opensource`）**，依赖 **MySQL、Redis**。

## 前置条件

- 已安装 [Docker](https://docs.docker.com/get-docker/) 与 Docker Compose v2
- 构建阶段需从 Oracle 官网下载 **Instant Client Basic Lite**（`Dockerfile` 中 `wget`）；若企业网络拦截，请自行下载 zip 放到构建上下文并在 Dockerfile 中改为 `COPY` 本地文件后 `unzip`

## 一键启动（推荐）

在仓库 **`dbmeta-core` 根目录**执行：

```bash
docker compose -f docker/docker-compose.yml up -d --build
```

浏览器访问：<http://127.0.0.1:8086/>

- 默认数据库账号与 `docker/setting.docker.yml` 中 `dataSource` 一致（用户/库名均为 `dbmeta`，密码 `dbmeta`）
- 首次启动由程序自动建表；管理员账号由种子数据写入（见 `src/database/db.go` 中 `Users` 初始化）
- **生产环境**请修改 `setting.docker.yml` 中的密码、`token.key`、`decrypt` 密钥，并将 `license.devSkip` 设为 `false` 且按说明配置授权

## 仅构建镜像（不启动编排）

```bash
docker build -f docker/Dockerfile -t dbmeta-core:latest .
```

## 端口说明

| 服务    | 默认端口 |
|---------|----------|
| Web/API | 8086     |
| MySQL   | 3306     |
| Redis   | 6379     |

## 配置说明

- 容器内监听 `:8086`，与 `setting.docker.yml` 中 `server.addr` 一致
- 修改配置可编辑 `docker/setting.docker.yml` 后执行 `docker compose ... up -d` 重启 `app` 服务
