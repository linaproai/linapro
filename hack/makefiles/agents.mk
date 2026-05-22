# Agents resource symlink management targets.
# Agents 资源软链管理目标。
#
# This Makefile fragment exposes two layered entry points:
#
# Support both the historical upper-case Make variables and the lower-case
# names used by linactl's key=value arguments.
agents.agent := $(or $(agent),$(AGENT))
agents.action := $(or $(action),$(ACTION))
agents.force := $(or $(force),$(FORCE))

#   1. `agents` (recommended) — agent-first one-shot/interactive setup.
#      - On a TTY without agent/AGENT, opens an arrow-key driven menu that
#        first picks the agent, then picks link or unlink. The chosen
#        action is automatically applied to every resource type
#        (skills / prompts / md) the agent participates in; resources
#        where the agent is native or unregistered are skipped with an
#        explicit reason in the final summary.
#      - With agent=<name> (or AGENT=<name>), runs the same dispatch
#        non-interactively.
#        action defaults to `link`; pass action=unlink to remove.
#        agent must be a single supported agent name (no `all`, no
#        comma-separated list). Upper-case AGENT/ACTION/FORCE remain
#        compatibility aliases.
#
#   2. `agents.<resource>.<action>` (advanced) — per-resource batch
#      operations preserved from before:
#        - skills:  directory bridge from .<tool>/skills    -> .agents/skills
#        - prompts: directory bridge from .<tool>/commands  -> .agents/prompts
#        - md:      single-file bridge from .<tool>.md      -> AGENTS.md
#      These accept agent=<name|all|csv> (or AGENT=<name|all|csv>) and
#      remain the recommended route for batch updates across many agents
#      at once.

.PHONY: agents \
        agents.skills.link agents.skills.unlink \
        agents.prompts.link agents.prompts.unlink \
        agents.md.link agents.md.unlink

# agents drives the agent-first aggregate command. Without arguments and
# attached to a TTY, it opens the arrow-key picker. With agent/AGENT set, it
# runs non-interactively against every resource the agent participates
# in. Pass force=1 to rebuild mismatched links, action=unlink to remove.
agents:
	@$(LINACTL) agents $(if $(agents.agent),agent=$(agents.agent)) $(if $(agents.action),action=$(agents.action)) $(if $(agents.force),force=1)

# agents.skills.link manages repository-local symlinks from supported
# agents' project skills paths to .agents/skills. Pass agent=<name|all|csv>
# to create or rebuild links; pass force=1 to rebuild mismatched links or
# enable rootCollision agents.
agents.skills.link:
	@$(LINACTL) agents.skills.link $(if $(agents.agent),agent=$(agents.agent)) $(if $(agents.force),force=1)

# agents.skills.unlink removes repository-local skills symlinks managed
# by agents.skills.link. It never removes real directories or files.
# Pass agent=<name|all|csv>.
agents.skills.unlink:
	@$(LINACTL) agents.skills.unlink $(if $(agents.agent),agent=$(agents.agent))

# agents.prompts.link manages repository-local symlinks from supported
# agents' commands/prompts roots to .agents/prompts. Pass
# agent=<name|all|csv> and optional force=1.
agents.prompts.link:
	@$(LINACTL) agents.prompts.link $(if $(agents.agent),agent=$(agents.agent)) $(if $(agents.force),force=1)

# agents.prompts.unlink removes repository-local prompts symlinks managed
# by agents.prompts.link.
agents.prompts.unlink:
	@$(LINACTL) agents.prompts.unlink $(if $(agents.agent),agent=$(agents.agent))

# agents.md.link manages repository-local symlinks from supported agents'
# private project guide files (e.g. CLAUDE.md, GEMINI.md) to AGENTS.md.
# Pass agent=<name|all|csv> and optional force=1.
agents.md.link:
	@$(LINACTL) agents.md.link $(if $(agents.agent),agent=$(agents.agent)) $(if $(agents.force),force=1)

# agents.md.unlink removes repository-local AGENTS.md symlinks managed
# by agents.md.link. Real authored files (e.g. a hand-written CLAUDE.md)
# are never removed.
agents.md.unlink:
	@$(LINACTL) agents.md.unlink $(if $(agents.agent),agent=$(agents.agent))
