# Tier Classification

This document is the canonical path classification source for the `lina-upgrade` skill.
`scripts/upgrade-classify.sh` implements the same rules in executable form.

## Tier Table

| Tier | Path patterns | Reason |
| --- | --- | --- |
| Tier 1 | `apps/lina-core/pkg/bizerr/**` | Stable business-error contract shared by host, plugins, and API responses. |
| Tier 1 | `apps/lina-core/pkg/logger/**` | Stable logging facade used by host and plugin backend code. |
| Tier 1 | `apps/lina-core/pkg/contract/**` | Public contract package. |
| Tier 1 | `apps/lina-core/pkg/pluginbridge/**`, `pkg/pluginhost/**`, `pkg/pluginservice/**`, `pkg/plugincontroller/**`, `pkg/pluginfs/**`, `pkg/plugindb/**` | Public plugin runtime API surfaces. |
| Tier 1 | `apps/lina-core/pkg/sourceupgrade/contract/**` | Public source-upgrade governance contract. |
| Tier 1 | `apps/lina-plugins/<plugin-id>/**` except generated paths | User-owned or first-party source-plugin directories. |
| Tier 2 | `apps/lina-core/internal/**` except generated paths | User-modifiable host implementation. |
| Tier 2 | `apps/lina-vben/apps/web-antd/src/**` except generated paths | User-modifiable default workspace implementation. |
| Tier 2 | `apps/lina-core/manifest/config/*.yaml` except the `framework.version` field | User-modifiable runtime configuration. |
| Tier 2 | Unknown files under `apps/lina-core/`, `apps/lina-vben/`, or `apps/lina-plugins/` | Safest fallback because semantic review may be required. |
| Tier 3 | `apps/lina-core/internal/dao/**` | Generated `DAO` artifacts. |
| Tier 3 | `apps/lina-core/internal/model/do/**` | Generated `DO` artifacts. |
| Tier 3 | `apps/lina-core/internal/model/entity/**` | Generated entity artifacts. |
| Tier 3 | `apps/lina-core/internal/controller/**` | GoFrame-generated controller skeletons. |
| Tier 3 | `apps/lina-plugins/<plugin-id>/backend/internal/dao/**` | Plugin-local generated `DAO` artifacts. |
| Tier 3 | `apps/lina-plugins/<plugin-id>/backend/internal/model/do/**` | Plugin-local generated `DO` artifacts. |
| Tier 3 | `apps/lina-plugins/<plugin-id>/backend/internal/model/entity/**` | Plugin-local generated entity artifacts. |
| Tier 3 | `apps/lina-plugins/<plugin-id>/backend/internal/controller/**` | Plugin-local generated controller skeletons. |

## Fallback

When a path is not matched by any explicit rule, classify it as `tier2` if it is under `apps/lina-core/`, `apps/lina-vben/`, or `apps/lina-plugins/`.
Paths outside those roots return `unknown`; the AI must inspect them before taking action.

## Relationship to `upgrade-classify.sh`

The script is deliberately simple and path-based. When this document changes, update `scripts/upgrade-classify.sh` and its tests in the same change.
