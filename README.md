# Lina

Lina is an AI-driven full-stack development framework. It combines a GoFrame-based core host service, a Vue 3 + Vben 5 based default management workspace, plugin extensibility, and an OpenSpec-powered AI collaboration workflow.

## Highlights

- `apps/lina-core`: the core host service that exposes reusable module APIs, shared platform capabilities, governance services, and plugin runtime support.
- `apps/lina-vben`: the default management workspace used as the reference front-end application.
- `apps/lina-plugins`: source-plugin and dynamic-plugin samples used as implementation references.
- `openspec/`: change proposals, designs, specs, and task plans for structured delivery.
- `hack/tests`: Playwright E2E coverage for user-visible behavior.

## Repository Layout

```text
apps/
  lina-core/      Core host service (GoFrame)
  lina-vben/      Default management workspace (Vue 3 + Vben 5)
  lina-plugins/   Plugin samples and plugin development references
hack/
  tests/          Playwright E2E suite
openspec/
  changes/        Active and archived OpenSpec changes
  specs/          Current baseline capability specs
```

## Getting Started

### Prerequisites

- Go
- Node.js
- pnpm
- MySQL

### Common Commands

```bash
make dev          # Start backend and frontend
make stop         # Stop local services
make status       # Show local service status
make init         # Apply host SQL and seed data
make mock         # Load mock/demo data
make test         # Run the Playwright suite
```

Backend development:

```bash
cd apps/lina-core
go run main.go
make build
make ctrl
make dao
```

Front-end development:

```bash
cd apps/lina-vben
pnpm install
pnpm -F @lina/web-antd dev
pnpm run build
```

## Default Account

- Username: `admin`
- Password: `admin123`

## Delivery Workflow

Lina uses OpenSpec as the structured delivery backbone.

1. Explore the requirement and solution space.
2. Create an OpenSpec proposal under `openspec/changes/`.
3. Implement tasks incrementally.
4. Update tests and documentation alongside code.
5. Review, verify, and archive after acceptance.

## Documentation Policy

- Every directory-level primary document uses English `README.md`.
- Every directory that has `README.md` must also provide a synchronized Chinese mirror named `README.zh_CN.md`.
- The two files must keep the same structure and technical facts.

## Entry Points

- `CLAUDE.md`: repository-wide engineering rules and workflow guidance.
- `apps/lina-plugins/README.md`: plugin system overview.
- `openspec/specs/`: current capability baselines.
