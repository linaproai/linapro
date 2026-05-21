# Agents resource symlink management targets.
# Agents 资源软链管理目标。
#
# This Makefile fragment exposes two layered entry points:
#
#   1. `agents` (recommended) — agent-first one-shot/interactive setup.
#      - On a TTY without AGENT, opens an arrow-key driven menu that
#        first picks the agent, then picks link or unlink. The chosen
#        action is automatically applied to every resource type
#        (skills / prompts / md) the agent participates in; resources
#        where the agent is native or unregistered are skipped with an
#        explicit reason in the final summary.
#      - With AGENT=<name>, runs the same dispatch non-interactively.
#        ACTION defaults to `link`; pass ACTION=unlink to remove.
#        AGENT must be a single supported agent name (no `all`, no
#        comma-separated list).
#
#   2. `agents.<resource>.<action>` (advanced) — per-resource batch
#      operations preserved from before:
#        - skills:  directory bridge from .<tool>/skills    -> .agents/skills
#        - prompts: directory bridge from .<tool>/.../opsx  -> .agents/prompts/opsx
#        - md:      single-file bridge from .<tool>.md      -> AGENTS.md
#      These accept AGENT=<name|all|csv> and remain the recommended
#      route for batch updates across many agents at once.

.PHONY: agents \
        agents.skills.link agents.skills.unlink \
        agents.prompts.link agents.prompts.unlink \
        agents.md.link agents.md.unlink

# agents drives the agent-first aggregate command. Without arguments and
# attached to a TTY, it opens the arrow-key picker. With AGENT set, it
# runs non-interactively against every resource the agent participates
# in. Pass FORCE=1 to rebuild mismatched links, ACTION=unlink to remove.
agents:
	$(LINACTL) agents $(if $(AGENT),agent=$(AGENT)) $(if $(ACTION),action=$(ACTION)) $(if $(FORCE),force=1)

# agents.skills.link manages repository-local symlinks from supported
# agents' project skills paths to .agents/skills. Pass AGENT=<name|all|csv>
# to create or rebuild links; pass FORCE=1 to rebuild mismatched links or
# enable rootCollision agents.
agents.skills.link:
	$(LINACTL) agents.skills.link $(if $(AGENT),agent=$(AGENT)) $(if $(FORCE),force=1)

# agents.skills.unlink removes repository-local skills symlinks managed
# by agents.skills.link. It never removes real directories or files.
# Pass AGENT=<name|all|csv>.
agents.skills.unlink:
	$(LINACTL) agents.skills.unlink $(if $(AGENT),agent=$(AGENT))

# agents.prompts.link manages repository-local symlinks from supported
# agents' commands/prompts paths to per-agent source directories under
# .agents/prompts/. Pass AGENT=<name|all|csv> and optional FORCE=1.
agents.prompts.link:
	$(LINACTL) agents.prompts.link $(if $(AGENT),agent=$(AGENT)) $(if $(FORCE),force=1)

# agents.prompts.unlink removes repository-local prompts symlinks managed
# by agents.prompts.link.
agents.prompts.unlink:
	$(LINACTL) agents.prompts.unlink $(if $(AGENT),agent=$(AGENT))

# agents.md.link manages repository-local symlinks from supported agents'
# private project guide files (e.g. CLAUDE.md, GEMINI.md) to AGENTS.md.
# Pass AGENT=<name|all|csv> and optional FORCE=1.
agents.md.link:
	$(LINACTL) agents.md.link $(if $(AGENT),agent=$(AGENT)) $(if $(FORCE),force=1)

# agents.md.unlink removes repository-local AGENTS.md symlinks managed
# by agents.md.link. Real authored files (e.g. a hand-written CLAUDE.md)
# are never removed.
agents.md.unlink:
	$(LINACTL) agents.md.unlink $(if $(AGENT),agent=$(AGENT))
