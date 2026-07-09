# LinaPro plugin code generation targets.
# Plugin Makefiles may set PLUGIN_ROOT before including this shared fragment.

PLUGIN_ROOT ?= $(CURDIR)
REPO_ROOT ?= $(abspath $(PLUGIN_ROOT)/../../..)
PLUGIN_BACKEND ?= $(PLUGIN_ROOT)/backend
LINACTL ?= go -C "$(REPO_ROOT)/hack/tools/linactl" run .

# ctrl: Generate this plugin's GoFrame controller scaffolding
# ctrl: 生成当前插件的 GoFrame 控制器骨架
.PHONY: ctrl
ctrl:
	@$(LINACTL) ctrl dir="$(PLUGIN_BACKEND)"

# dao: Generate this plugin's DAO/DO/Entity files
# dao: 生成当前插件的 DAO/DO/Entity 文件
.PHONY: dao
dao:
	@$(LINACTL) dao dir="$(PLUGIN_BACKEND)"
