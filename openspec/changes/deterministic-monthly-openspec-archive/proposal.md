# Deterministic Monthly OpenSpec Archive

The monthly OpenSpec archive workflow currently asks an AI tool to perform the base archive step. The failed run `26166669462` showed that the AI process can exit successfully while leaving completed active changes unarchived, so the workflow fails later in the assertion step without producing a partial archive PR.

This change keeps archive execution inside the configured AI coding tool workflow. Codex, Claude Code, and GitHub Copilot CLI each run the same auto-archive prompt through their own runtime before running the shared archive consolidation prompt through that same runtime. The workflow must surface unarchived completed changes with clear failure details and stop immediately when any archive, consolidation, validation, push, or pull request step fails.

The same change fixes the known `remove-sqlite-support` archive blocker by aligning its OpenSpec delta with the current `cluster-coordination-config` baseline, and updates artifact upload actions away from the Node 20 runtime generation.
