# LinaPro Plugin Workspace Commands
# LinaPro 插件工作区目标
# =================

PLUGIN_CONFIG_ARGS :=
ifneq ($(origin config), undefined)
PLUGIN_CONFIG_ARGS += --config=$(config)
endif
ifneq ($(origin p), undefined)
PLUGIN_CONFIG_ARGS += p=$(p)
endif
ifneq ($(origin source), undefined)
PLUGIN_CONFIG_ARGS += source=$(source)
endif
ifneq ($(origin force), undefined)
PLUGIN_CONFIG_ARGS += force=$(force)
endif

# Convert apps/lina-plugins from a submodule into a normal directory.
# 将 apps/lina-plugins 从 submodule 转换为普通目录。
## plugins.init: Convert apps/lina-plugins from submodule to ordinary directory while preserving plugin code
.PHONY: plugins.init
plugins.init:
	@$(LINACTL) plugins.init $(PLUGIN_CONFIG_ARGS)

# Install configured source plugins into apps/lina-plugins.
# 安装配置中的源码插件到 apps/lina-plugins。
## plugins.install: Install configured plugins from hack/config.yaml; use p=<plugin-id> or source=<name> for a subset
.PHONY: plugins.install
plugins.install:
	@$(LINACTL) plugins.install $(PLUGIN_CONFIG_ARGS)

# Update configured source plugins in apps/lina-plugins.
# 更新 apps/lina-plugins 中配置的源码插件。
## plugins.update: Update configured plugins from hack/config.yaml; dirty plugins are blocked unless force=1
.PHONY: plugins.update
plugins.update:
	@$(LINACTL) plugins.update $(PLUGIN_CONFIG_ARGS)

# Show configured source-plugin workspace status.
# 查看配置化源码插件工作区状态。
## plugins.status: Show plugin workspace status, configured plugin versions, local changes, and remote update state
.PHONY: plugins.status
plugins.status:
	@$(LINACTL) plugins.status $(PLUGIN_CONFIG_ARGS)
