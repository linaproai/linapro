# Upgrade Source Tool

`upgrade-source` is the development-only upgrade command behind the repository `make upgrade` target. It supports framework upgrades and source-plugin upgrade governance.

## Usage

Preferred repository entry point:

```bash
make upgrade confirm=upgrade
make upgrade confirm=upgrade scope=framework target=<tag-or-ref>
make upgrade confirm=upgrade scope=source-plugin plugin=<plugin-id>
make upgrade confirm=upgrade scope=source-plugin plugin=all
make upgrade confirm=upgrade scope=source-plugin plugin=all dry_run=1
```

Direct tool invocation:

```bash
go run ./hack/tools/upgrade-source --confirm=upgrade --scope=framework --target=<tag-or-ref>
go run ./hack/tools/upgrade-source --confirm=upgrade --scope=source-plugin --plugin=<plugin-id>
go run ./hack/tools/upgrade-source --confirm=upgrade --scope=source-plugin --plugin=all --dry-run
```

## Options

| Option | Scope | Description |
| --- | --- | --- |
| `--confirm=upgrade` | All | Required confirmation token for upgrade and database-sensitive operations. |
| `--scope` | All | Upgrade scope. Supported values: `framework`, `source-plugin`. Defaults to `framework`. |
| `--repo` | `framework` | Optional upstream framework Git repository URL. Defaults to the repository configured in `apps/lina-core/hack/config.yaml`. |
| `--target` | `framework` | Optional target framework tag or Git reference. |
| `--plugin` | `source-plugin` | Source plugin ID, or `all` for every source plugin. |
| `--dry-run` | All | Prints the resolved plan without applying code, SQL, or source-plugin governance changes. |

## Behavior

- Framework upgrade mode verifies the workspace, resolves the target framework release, prints the upgrade plan, synchronizes code, and replays host SQL files when needed.
- Source-plugin mode lists discovered source plugin versions, compares them with effective host versions, rejects downgrade cases, and applies the prepared source-plugin release when execution is allowed.
- The command refuses to run unless `--confirm=upgrade` is present.

## Notes

- Commit or stash local changes before running framework upgrades. The tool performs a clean working-tree precheck.
- Back up the repository and database before running a real upgrade.
- Use `--dry-run` first when checking a framework target or source-plugin batch.
