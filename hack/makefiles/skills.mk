# Skill link management targets.
# 技能软链接管理目标。

.PHONY: skills skills.link skills.unlink

# skills opens an interactive action menu (link / unlink) when invoked on a
# TTY. CI and piped contexts print usage guidance pointing at the explicit
# subcommands instead.
skills:
	$(LINACTL) skills

# skills.link manages repository-local symlinks from supported agents' project
# skill paths to .agents/skills. Pass AGENT=<name|all|csv> to create or rebuild
# links; pass FORCE=1 to rebuild mismatched links or enable rootCollision agents.
skills.link:
	$(LINACTL) skills.link $(if $(AGENT),agent=$(AGENT)) $(if $(FORCE),force=1)

# skills.unlink removes repository-local symlinks managed by skills.link.
# It never removes real directories or files. Pass AGENT=<name|all|csv>.
skills.unlink:
	$(LINACTL) skills.unlink $(if $(AGENT),agent=$(AGENT))
