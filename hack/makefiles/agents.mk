# Agents resource symlink management targets.
# Agents 资源软链管理目标。
#
# This Makefile fragment exposes the agents.<resource>.<action> command
# tree provided by linactl. Three resource types are supported, each
# with a link/unlink action pair:
#   - skills:  directory bridge from .<tool>/skills    -> .agents/skills
#   - prompts: directory bridge from .<tool>/.../opsx  -> .agents/prompts/opsx
#   - md:      single-file bridge from .<tool>.md      -> AGENTS.md
#
# The bare `agents` target opens an interactive resource/action menu on
# a TTY and prints usage guidance otherwise.

.PHONY: agents \
        agents.skills.link agents.skills.unlink \
        agents.prompts.link agents.prompts.unlink \
        agents.md.link agents.md.unlink

# agents opens an interactive three-level menu (resource -> action ->
# agent) when invoked on a TTY. CI and piped contexts print usage
# guidance pointing at the explicit subcommands instead.
agents:
	$(LINACTL) agents

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
