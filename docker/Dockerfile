# syntax=docker/dockerfile:1
# 开源版一体化镜像：构建 Ant Design 前端（嵌入 /public/）+ Go 后端（-tags opensource）
# 构建上下文：仓库根目录（dbmeta-core）
#   docker build -f docker/Dockerfile -t dbmeta-core:latest .

# ---------- 前端：Vite 产物写入 webassets/static，index 在 webassets 根目录 ----------
FROM node:22-bookworm AS frontend
WORKDIR /src/frontend
ENV CI=true
ENV NODE_OPTIONS=--max-old-space-size=8192
# 与 Gin 中 r.StaticFS("/public/", webassets.Static) 对齐：静态资源前缀 /public/
ENV VITE_BASE=/public/
ENV VITE_GLOB_API_URL=/api
RUN corepack enable && corepack prepare pnpm@10.33.0 --activate
COPY frontend/ .
RUN pnpm install --frozen-lockfile
RUN pnpm run build:antd

# ---------- Oracle Instant Client（godror 依赖 CGO + oci）----------
FROM debian:bookworm-slim AS oci
RUN apt-get update && apt-get install -y --no-install-recommends wget unzip ca-certificates \
  && rm -rf /var/lib/apt/lists/*
# Oracle 若变更下载地址，可改下列 URL 或改用本地 COPY
RUN mkdir -p /opt/oracle \
  && wget -qO /tmp/ic.zip \
    https://download.oracle.com/otn_software/linux/instantclient/219000/instantclient-basiclite-linux.x64-21.9.0.0.0dbru.zip \
  && unzip -q /tmp/ic.zip -d /opt/oracle \
  && rm /tmp/ic.zip \
  && IC=$(ls -d /opt/oracle/instantclient_* | head -1) \
  && ln -sfn "$IC" /opt/oracle/instantclient

# ---------- Go 编译 ----------
FROM golang:1.24-bookworm AS gobuild
RUN apt-get update && apt-get install -y --no-install-recommends gcc libc6-dev libaio1 \
  && rm -rf /var/lib/apt/lists/*
COPY --from=oci /opt/oracle /opt/oracle
ENV CGO_ENABLED=1
ENV GO111MODULE=on
ENV GOFLAGS=-mod=mod
ENV LD_LIBRARY_PATH=/opt/oracle/instantclient
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /src/frontend/apps/web-antd/dist /tmp/web-dist
# 清空旧版静态资源，避免与 Vite 产物混用
RUN rm -rf webassets/static \
  && mkdir -p webassets/static \
  && cp -a /tmp/web-dist/. webassets/static/ \
  && mv webassets/static/index.html webassets/index.html
RUN go build -tags opensource -trimpath -ldflags="-s -w" -o /out/dbmeta-core .

# ---------- 运行镜像 ----------
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends libaio1 ca-certificates tzdata \
  && rm -rf /var/lib/apt/lists/*
COPY --from=oci /opt/oracle /opt/oracle
ENV LD_LIBRARY_PATH=/opt/oracle/instantclient
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=gobuild /out/dbmeta-core /usr/local/bin/dbmeta-core
COPY docker/setting.docker.yml /app/setting.yml
EXPOSE 8088
CMD ["/usr/local/bin/dbmeta-core", "-c", "/app/setting.yml"]
