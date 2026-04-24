# 贡献指南（Contributing）

感谢你参与 DBMeta 项目贡献。本文档用于统一分支策略、提交规范与发布流程，确保代码质量与版本可追溯。

## 1. 分支模型

默认长期维护以下分支：

- `main`：稳定主干，始终保持可发布、可部署
- `feature/<scope>-<short-desc>`：新功能开发分支
- `fix/<scope>-<short-desc>`：缺陷修复分支
- `hotfix/<scope>-<short-desc>`：线上紧急修复分支（从 `main` 拉取）
- `release/vX.Y.Z`（可选）：发版准备分支

命名示例：

- `feature/meta-business-binding`
- `feature/capacity-growth-dashboard`
- `fix/docker-build-go-module`
- `hotfix/login-cookie-expired`
- `release/v1.0.0`

## 2. 开发与合并流程

1. 从 `main` 拉取最新代码
2. 创建功能分支（`feature/*` 或 `fix/*`）
3. 开发与自测
4. 提交 Pull Request 合并到 `main`
5. 通过评审与构建后合并

约束：

- 不直接向 `main` 提交代码
- 所有改动通过 PR 合并
- PR 必须填写变更说明、测试说明、风险与回滚方案

## 3. 提交信息建议（Conventional Commits）

推荐格式：

- `feat: ...` 新功能
- `fix: ...` 缺陷修复
- `refactor: ...` 重构
- `docs: ...` 文档
- `build: ...` 构建/依赖
- `chore: ...` 杂项维护

示例：

- `feat(meta): add database-business batch sync`
- `fix(docker): resolve go module build path`
- `docs(readme): update install and deploy steps`

## 4. 版本号规范（SemVer）

版本号采用 `vX.Y.Z`：

- `X`（Major）：不兼容变更
- `Y`（Minor）：向后兼容的新功能
- `Z`（Patch）：向后兼容的问题修复

## 5. 发布清单（Release Checklist）

每次发布前请完成以下检查：

- [ ] `main` 分支构建通过（后端 / 前端 / Docker）
- [ ] `CHANGELOG.md` 已从 `Unreleased` 归档到目标版本
- [ ] `VERSION` 已更新为目标版本（如 `1.0.0`）
- [ ] 关键功能已回归验证（登录、查询、任务、质量、容量等）
- [ ] 若包含 `webassets` 变更，已在 PR 说明对应前端源码范围与构建命令
- [ ] 打 Git Tag：`vX.Y.Z`
- [ ] 创建 GitHub Release 并附发布说明

发布命令示例：

```bash
git checkout main
git pull
git tag v1.0.0
git push origin v1.0.0
```

## 6. `webassets` 提交规则（B 方案）

本项目采用“`webassets` 产物可入库”策略：

- 允许提交：`webassets/index.html`、`webassets/static/**`
- 非前端发布类改动（纯后端/文档）不应携带 `webassets` 变更
- 若 PR 包含 `webassets` 变更，需说明：
  - 对应前端源码改动范围
  - 使用的构建命令（例如 `pnpm run build:antd`）
  - 变更原因（发版同步 / 功能上线 / 依赖升级）

## 7. 热修复流程（Hotfix）

1. 从 `main` 切 `hotfix/*`
2. 修复并提 PR 到 `main`
3. 合并后立即打补丁版本 Tag（如 `v1.0.1`）
4. 同步更新 `CHANGELOG.md`（`Fixed`）

## 8. 维护建议

- 建议启用 `main` 保护规则（至少要求 PR 合并）
- 建议启用最小 CI（`go build`、前端 build、Docker build）
- 建议每个版本保持“可回滚、可追溯、可复现”
