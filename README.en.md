# DBMeta · Open Source Data Governance Platform

[中文](./README.md) | English | [Changelog](./CHANGELOG.md)

DBMeta is an open-source platform for database governance, providing unified capabilities across metadata management, data quality governance, task orchestration, and AI-assisted analysis.

This repository is the **core open-source repository**, including backend services, frontend projects, and deployment assets, suitable for self-hosted enterprise usage and secondary development.

---

## Product Positioning

DBMeta is designed to solve three key challenges in database governance:

- **Low asset visibility**: data sources, schemas, tables, fields, and business context are spread across systems, with no unified view.
- **Broken governance loop**: quality rules, scheduled tasks, and issue tracking are disconnected and hard to operate continuously.
- **High analysis barrier**: SQL writing, model integration, and cross-role collaboration are expensive.

By combining a governance model, task system, and AI capabilities, DBMeta builds a sustainable and evolvable governance workspace.

---

## Key Highlights

### 1) End-to-end governance loop
- Covers core governance objects: data sources, instances, databases, tables, fields, and business metadata.
- Provides quality rules, quality tasks, issue tracking, and governance dashboards.
- Supports evolution from configuration-driven governance to operation-driven governance.

### 2) Ready-to-use task system
- Supports both scheduled and manual execution.
- Full visibility for task status, execution logs, and results.
- Extensible task model for custom governance workflows.

### 3) AI-enhanced analytics experience
- Includes chat, rules, sessions, and model management.
- Supports multi-model configuration and routing strategies.
- Embeds AI into governance workflows instead of isolated chat tooling.

### 4) Full-stack delivery
- Backend and frontend work together with same-origin access.
- Provides one-command Docker deployment path.
- Clear project structure for team collaboration and extensibility.

---

## Capability Modules

| Module | Capability |
|---|---|
| Metadata Governance | Data source management, schema/table/field management, business metadata maintenance |
| Data Query | Query entry, favorites, and permission boundaries |
| Data Quality | Rules, tasks, issues, and dashboards |
| AI Capabilities | AI chat, rules, model configuration, and session management |
| Capacity Analytics | Capacity statistics, growth analysis, and Top-N views |
| System Tasks | Scheduling, task logs, and task options |

---

## Project Structure

```text
dbmeta-core/
├─ app/                 # bootstrap and startup
├─ router/              # route registration
├─ setting/             # config parsing
├─ src/
│  ├─ controller/       # HTTP controllers
│  ├─ service/          # business service layer
│  ├─ model/            # data models
│  ├─ database/         # db initialization and migrations
│  ├─ task/             # scheduled/background tasks
│  ├─ module/           # module registry and extension points
│  └─ ...
├─ frontend/            # frontend monorepo (Vben-based)
├─ webassets/           # embedded static assets
└─ docker/              # Docker deployment files
```

---

## Quick Start

### 1) Requirements

- Go 1.19+
- MySQL 8+
- Redis 6+
- Node.js / pnpm (for frontend development)

### 2) Run locally (backend)

```bash
go mod tidy
go run . -c ./setting.yml
```

> It is recommended to copy `setting.example.yml` first, then fill in local connection settings.

### 3) One-command Docker deployment

```bash
docker compose -f docker/docker-compose.yml up -d --build
```

Default access URL: `http://127.0.0.1:8086`

---

## Configuration

- Example config: `setting.example.yml`
- Local config: `setting.yml` (already ignored by `.gitignore`)
- Notification-related business config has been moved to database settings tables
- In production, use isolated config files and proper secret management

---

## Development Guide

### Backend
- Entry point: `main.go`
- Bootstrap: `app/bootstrap.go`
- Routes: `router/router.go`

### Frontend
- Frontend project is under `frontend/`
- Refer to subproject docs/scripts for local frontend startup

### Build check

```bash
go build .
```

### Frontend Build Artifacts Policy (Plan B)

- `webassets/` is kept in the repository as built frontend artifacts to support backend-first, out-of-box startup.
- Allowed tracked artifacts: `webassets/index.html` and `webassets/static/**`.
- Non-frontend-release changes (for example backend-only, docs-only, or scripts-only) should not include `webassets` diffs.
- If a PR includes `webassets` changes, describe the related frontend source scope and build command in the PR body.
- It is recommended to update `webassets` only for release sync, frontend feature delivery, or frontend dependency upgrades.

---

## FAQ

### Port conflict on startup
Set another `server.addr` in config and restart.

### Frontend API target is wrong
Check frontend proxy configuration and confirm the backend target address.

### Blank page after startup
Make sure `webassets` static resources match the current backend build.

---

## Contributing

Contributions via Issue / Pull Request are welcome:

- Bug fixes
- Documentation improvements
- Governance rule/task extensions
- AI capability and model integration improvements

Please run basic build checks and include test notes before submitting.

---

## License

AGPL-3.0 for non-commercial use. Commercial license required for any commercial use.

| Use Case | Allowed? |
|---|---|
| Personal / research / educational | Yes |
| Self-hosted (non-commercial) | Yes, with attribution |
| Fork and modify (non-commercial) | Yes, share source under AGPL-3.0 |
| Commercial use / SaaS / rebranding | Requires commercial license |

See LICENSE for full terms. For commercial licensing, contact the maintainer.

Copyright (C) 2026 DBMETA.COM  All rights reserved.

