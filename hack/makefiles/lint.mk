# LinaPro Lint Commands
# LinaPro 静态检查指令
# =====================

LINT_GO_ARGS :=
ifneq ($(origin plugins), undefined)
LINT_GO_ARGS += plugins=$(plugins)
endif
ifneq ($(origin fix), undefined)
LINT_GO_ARGS += fix=$(fix)
endif

# Run the Go static lint gate for every selected workspace module.
# 对选定 Go workspace 模块运行 Go 静态检查门禁。
## lint.go: Run Go static lint checks
.PHONY: lint.go
lint.go:
	@$(LINACTL) lint.go $(LINT_GO_ARGS)

# Run the repository lint gate. It currently delegates to Go static lint.
# 运行仓库静态检查门禁；当前转发到 Go 静态检查。
## lint: Run repository lint checks
.PHONY: lint
lint: lint.go
