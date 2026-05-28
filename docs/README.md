# LinaPro Backend Documentation

This directory hosts cross-cutting technical guides for the LinaPro backend
and the source-plugin ecosystem. Documents here are intentionally generic:
host-internal architecture notes live next to their code (in package
`README` or doc comments), while this directory is reserved for guides
that span multiple packages or describe an integration contract.

> 中文镜像见 [README.zh-CN.md](./README.zh-CN.md).

## Index

| Document | Topic |
|---|---|
| [auth-provider-integration.md](./auth-provider-integration.md) | How to add a new third-party authentication provider (Google, Discord, GitHub, custom OIDC, …) as a source plugin. |

## Conventions

- Every guide ships an English source (`<name>.md`) and a Chinese mirror
  (`<name>.zh-CN.md`). The two must stay content-equivalent — when you
  change one, change the other in the same commit.
- Code samples are illustrative, not copy-paste. The canonical patterns
  live in the corresponding repository code (typically under
  `apps/lina-core/pkg/...` or `apps/lina-plugins/<plugin-id>/...`).
- When a guide references a public API surface, it must cite the file
  path that owns the contract so readers can verify behavior at the
  source of truth.
