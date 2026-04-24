# 变更日志

本文档记录 DBMeta 的主要版本变更。

格式参考 Keep a Changelog，版本号遵循 SemVer。

## [Unreleased]

### Added

- 新增项目级 `AGPL-3.0` 许可证文本（根目录 `LICENSE`）。
- README 增加 Star History 展示，并切换为本项目仓库数据源。

### Changed

- `README.md` 结构重排，采用更清晰的分区组织（项目概览、核心能力、Quick Start、Tech Stack、License 等）。
- 中英文 README 的 License 文案与商用说明对齐，统一为当前授权策略。

### Docs

- 新增中文变更日志文件 `CHANGELOG.md`，用于后续版本持续记录。
- 新增仓库级 PR 模板（`.github/pull_request_template.md`），补充 `webassets`（B 方案）变更检查项，并同步更新中英文 README 的前端产物管理规范。

## [1.0.0-rc.1] - 2026-04-24

### Added

- 发布 DBMeta 首个公开候选版本（RC）。
- 提供数据库治理核心能力：元数据治理、数据质量治理、任务编排与 AI 辅助分析。
- 提供前后端一体化工程与 Docker 部署路径。

### Changed

- 持续优化系统结构、权限控制与查询能力。
- 完善前端交互与部分模块的可用性表现。

### Fixed

- 修复若干日志、权限与流程中的稳定性问题。

---

> 说明：历史早期提交信息粒度有限，以上版本内容按模块进行归纳。后续版本建议在每次发布时同步更新本文件。
