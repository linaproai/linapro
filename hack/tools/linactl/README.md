# linactl

`linactl` is LinaPro's cross-platform development command entrypoint. It keeps the repository's long-lived task orchestration in Go so Windows, Linux, and macOS can run the same commands without depending on GNU Make or POSIX shell tools.

## Usage

```bash
cd hack/tools/linactl
go run . help
go run . status
go run . pack.assets
go run . wasm p=linapro-demo-dynamic
go run . wasm plugin_dir=/path/to/plugin out=temp/output
go run . plugins.status
go run . i18n.check
go run . init confirm=init
go run . tidy
go run . build platforms=linux/amd64,linux/arm64
go run . image tag=v0.2.0 push=0
go run . release.tag.check tag=v0.2.0
go run . release.tag.check print-version=1
```

## Windows Entry

The repository root also provides `make.cmd` as a thin Windows wrapper:

```cmd
make.cmd help
make.cmd status
make.cmd pack.assets
make.cmd plugins.status
make.cmd i18n.check
make.cmd init confirm=init
make.cmd tidy
make.cmd release.tag.check tag=v0.2.0
```

In PowerShell, run it with an explicit current-directory prefix:

```powershell
.\make.cmd help
.\make.cmd status
.\make.cmd pack.assets
.\make.cmd i18n.check
.\make.cmd release.tag.check tag=v0.2.0
```

## Parameters

`linactl` accepts the existing make-style `key=value` arguments to keep command migration low-friction.

| Parameter | Example | Purpose |
| --- | --- | --- |
| `confirm` | `confirm=init` | Confirms destructive bootstrap commands. |
| `rebuild` | `rebuild=true` | Rebuilds the configured database during `init`. |
| `platforms` | `platforms=linux/amd64,linux/arm64` | Selects build target platforms. |
| `plugins` | `plugins=0` | Overrides automatic plugin-full detection for build, dev, image, and Go test commands. |
| `tag` | `tag=v0.2.0` | Selects the release tag checked by `release.tag.check`. |
| `print-version` | `print-version=1` | Prints the validated `framework.version` for release automation. |
| `p` | `p=linapro-tenant-core` | Selects one plugin for Wasm build or plugin workspace management commands. |
| `plugin-dir` | `plugin_dir=/path/to/plugin` | Builds one dynamic plugin artifact from an explicit source directory. |
| `out` | `out=temp/output` | Selects the dynamic plugin artifact output directory. |
| `source` | `source=official` | Selects one configured plugin source for plugin workspace management commands. |
| `force` | `force=1` | Allows plugin install/update commands to overwrite existing or dirty plugin directories. |
| `verbose` | `verbose=1` | Shows child command output for build tasks. |

When `plugins` is omitted, build and dev commands enable plugin-full mode if `apps/lina-plugins` contains plugin manifests. Plugin-full mode generates or refreshes ignored `temp/go.work.plugins` from the host-only root `go.work`, then resolves source-plugin Go modules through `GOWORK`.

## Build Tool Commands

`linactl` owns the repository image build and dynamic plugin `Wasm` packaging implementation. The public entrypoints remain the root `make` targets and their direct `linactl` equivalents:

```bash
make image tag=v0.2.0 push=0
make image.build tag=v0.2.0
make wasm p=linapro-demo-dynamic
```

Use `plugin_dir=<path>` when a test or local fixture needs to package a dynamic plugin outside `apps/lina-plugins`.

## Runtime I18n Checks

`linactl i18n.check` owns the runtime `i18n` governance checks. It scans high-risk runtime-visible hard-coded copy and validates host/plugin runtime message key coverage:

```bash
make i18n.check
go run . i18n.check
```

The default scanner allowlist is maintained at `hack/tools/linactl/internal/runtimei18n/allowlist.json`.

## Agent Symlinks (agents.* command tree)

`linactl agents.<resource>.<action>` manages repository-local symlinks that bridge canonical sources under `.agents/` (and `AGENTS.md`) to per-agent project paths used by supported AI coding agents. Three resource types are supported:

- **skills** — directory bridge from `.<tool>/skills` to `.agents/skills`. The supported agent list mirrors [vercel-labs/skills](https://github.com/vercel-labs/skills#supported-agents).
- **prompts** — directory bridge from each agent's commands/prompts root (for example `.claude/commands`) to `.agents/prompts`.
- **md** — single-file bridge from `.<tool>.md` (or other private guide file) to the repo-root `AGENTS.md`.

The commands only operate inside the repository root; they never modify HOME directories or system-global paths, and they never remove real directories or files (even with `force=1`).

### Aggregate menu (recommended)

The `agents` aggregate command is **agent-first**: pick one agent and the chosen action runs across every resource type that agent participates in. Resources where the agent is `native` or unregistered are skipped with an explicit reason in the final summary.

```bash
# Interactive (TTY):
#   Step 1: arrow-key pick the agent (filter by typing).
#   Step 2: arrow-key pick `link` or `unlink`.
make agents

# One-shot (works in any environment, including CI):
make agents agent=claude-code                    # link claude-code across skills + prompts + md
make agents agent=ClaudeCode                     # same as agent=claude-code
make agents agent=claude-code force=1            # rebuild mismatched links during the same run
make agents agent=claude-code action=unlink      # remove every managed symlink for claude-code
```

`agent` must name a single supported agent: `agent=all` and comma-separated lists are explicitly rejected by the aggregate command (use the per-resource subcommands below for batch flows). Agent names are normalized to canonical kebab-case, so `ClaudeCode`, `Claude Code`, `claude_code`, and `claude-code` all resolve to `claude-code`. `action` defaults to `link`. Without `agent`, non-TTY invocations print a usage hint instead of blocking on input. The upper-case Make variables `AGENT`, `ACTION`, and `FORCE` remain accepted as compatibility aliases, but new examples should prefer the lower-case names because they match `linactl`'s `key=value` parameters.

### Per-resource subcommands (advanced)

The aggregate `make agents` command is the recommended entry point. The per-resource subcommands below remain available for batch flows that the aggregate command intentionally does not support, in particular `agent=all` and comma-separated lists.

```bash
# skills
make agents.skills.link                              # interactive selection on a TTY; read-only listing on CI/pipes
make agents.skills.link agent=claude-code            # create a single agent's link (non-interactive)
make agents.skills.link agent=claude-code,qoder      # create several agents' links
make agents.skills.link agent=all                    # create links for every link-class agent
make agents.skills.link agent=all force=1            # rebuild mismatched links
make agents.skills.unlink                            # interactive selection on a TTY (managed links only)
make agents.skills.unlink agent=claude-code          # remove one managed link
make agents.skills.unlink agent=all                  # remove every managed link

# prompts
make agents.prompts.link agent=claude-code           # link .claude/commands -> .agents/prompts
make agents.prompts.link agent=all                   # link every agent's commands/prompts directory
make agents.prompts.unlink agent=claude-code         # remove a managed prompts link

# md
make agents.md.link agent=claude-code                # link CLAUDE.md -> AGENTS.md
make agents.md.link agent=all                        # link every link-class agent's private guide file
make agents.md.unlink agent=claude-code              # remove a managed AGENTS.md link
```

### Interactive mode

All interactive entry points (the `agents` aggregate command and every `agents.<resource>.<action>` subcommand) are driven by [charmbracelet/huh](https://github.com/charmbracelet/huh): use the **arrow keys** to navigate, **space** to toggle multi-select rows, **enter** to confirm, **type** to filter, and **Esc** / **Ctrl+C** to cancel. CI and piped invocations remain non-interactive: `agents` prints a usage hint, `agents.<resource>.link` falls back to a read-only listing, and `agents.<resource>.unlink` requires an explicit `agent=` value.

Option labels follow two different conventions depending on the prompt:

- The aggregate `agents` command's "pick an agent" step is a single-select across the cross-resource registry. Each option shows only the human-readable agent name (for example `Claude Code`, `Codex`, `Cursor`) so the picker stays compact. The result table printed after confirmation lists which resources were applied or skipped.
- Per-resource subcommands (`agents.<resource>.<action>`) operate within a single resource and embed a **single-character status glyph** plus a short descriptor (e.g. `[~] claude-code  (mismatch)`) so you can see each candidate's current binding state without leaving the prompt.

Status glyphs embedded in per-resource option labels:

- `[+]` linked — symlink exists and points at the canonical source
- `[~]` mismatch — symlink exists but targets another location
- `[.]` absent — no symlink yet (or `native`, no action needed)
- `[!]` conflict — a real directory or file blocks linking
- `[*]` root-collision — agent uses a colliding repo-root path (only `openclaw` for skills)
- `[?]` error — inspection failed; see the non-interactive status table for details

### Categories

- `native` — agent reads the canonical source path directly. No symlink needed (e.g. for skills: `cursor`, `gemini-cli`, `codex`; for md: every agent that natively reads `AGENTS.md`).
- `link` — agent uses a different project path. A relative symlink to the canonical source is created on demand.
- `rootCollision` — project path is a bare repo-root name (only `skills/`, used by `openclaw`). Skipped by default; pass `agent=openclaw force=1` to opt in. Does not apply to prompts or md resources.

> **Fallback behaviour for `md`:** some agents auto-fall back to `AGENTS.md` when their preferred private guide file (e.g. `CODEBUDDY.md`, `CLAUDE.md`) is absent. CodeBuddy is one such agent — Tencent's docs state it prefers `CODEBUDDY.md` but loads `AGENTS.md` automatically when no `CODEBUDDY.md` is present. Agents with a documented automatic fallback are registered as `native` so cloned repositories work zero-config; agents whose preferred file is the *only* path they read are registered as `link` so you can opt into a symlink. See the inline comments in `internal/agents/md/md_agents.go` for the source-of-truth citation behind every entry.

Real directories or files at the target path are never auto-removed, even with `force=1`. `force=1` only rebuilds symlinks that already exist but point at a non-managed target. Per-tool skills and prompts symlinks are listed in `.gitignore`, so creating them locally does not pollute the repository.

### Migration from `make skills.*`

The old `make skills` / `make skills.link` / `make skills.unlink` targets and the corresponding `linactl skills*` subcommands have been **removed** in favor of the `agents.*` command tree. There are no aliases; existing scripts and documentation must be updated:

| Removed (no longer works) | Replacement |
| --- | --- |
| `make skills` | `make agents` |
| `make skills.link` | `make agents.skills.link` |
| `make skills.link AGENT=<name>` | `make agents.skills.link agent=<name>` |
| `make skills.link AGENT=all FORCE=1` | `make agents.skills.link agent=all force=1` |
| `make skills.unlink` | `make agents.skills.unlink` |
| `make skills.unlink AGENT=<name>` | `make agents.skills.unlink agent=<name>` |
| `linactl skills` | `linactl agents` |
| `linactl skills.link` | `linactl agents.skills.link` |
| `linactl skills.unlink` | `linactl agents.skills.unlink` |

The `agents.skills.*` subcommands behave identically to the previous `skills.*` commands (same registry, same status state machine, same TTY/CI behaviors). Only the command name changed.

## Release Tag Check

`release.tag.check` reads `apps/lina-core/manifest/config/metadata.yaml` and verifies that the release tag exactly matches `framework.version`.

```bash
make.cmd release.tag.check tag=v0.2.0
make release.tag.check tag=v0.2.0
make release.tag.check metadata=apps/lina-core/manifest/config/metadata.yaml tag=v0.2.0
```

In GitHub Actions, the command also accepts `GITHUB_REF_NAME` as the tag source when `tag` is omitted.

## Plugin Workspace Commands

Plugin workspace management always uses the fixed `apps/lina-plugins` directory. Configure sources in `hack/config.yaml`:

```yaml
plugins:
  sources:
    official:
      repo: "https://github.com/linaproai/official-plugins.git"
      root: "."
      ref: "main"
      items:
        - "linapro-tenant-core"
        - "linapro-org-core"
```

`items` only accepts plugin ID strings. Use the quoted string `"*"` to install every plugin directory directly under the source `root`; do not write bare `- *` because YAML treats it as alias syntax. If plugins from the same repository need different refs, split them into separate sources.

Common commands:

```bash
make plugins.init
make plugins.install
make plugins.install p=linapro-tenant-core
make plugins.update source=official
make plugins.update force=1
make plugins.status
```

`plugins.init` converts `apps/lina-plugins` from a submodule into a normal directory while preserving files. `plugins.install`, `plugins.update`, and `plugins.status` run the same workspace initialization automatically when needed, so users can start with the command they actually need. `plugins.install` and `plugins.update` reuse configured source checkouts under `temp/plugin-sources/<source>`, fetching updates after the first clone, copy plugin directories into `apps/lina-plugins/<plugin-id>`, and update the generated `apps/lina-plugins/.linapro-plugins.lock.yaml` lock file.

## Verification

```bash
cd hack/tools/linactl
go test ./...
go run . help
go run . wasm dry-run=true
go run . plugins.status
go run . i18n.check
go run . release.tag.check tag=v0.2.0
```
