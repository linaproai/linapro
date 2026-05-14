# LinaPro Test Commands
# LinaPro 测试指令
# ===================

TEST_GO_ARGS :=
ifneq ($(origin plugins), undefined)
TEST_GO_ARGS += plugins=$(plugins)
endif

# Run the complete Playwright E2E test suite.
# 运行完整的 Playwright E2E 测试套件。
## test: Run the full E2E test suite
.PHONY: test
test:
	@$(LINACTL) test scope="$(scope)"

# Run shell script unit and smoke tests for repository tooling.
# 运行仓库工具脚本的单元与 smoke 测试。
## test-scripts: Run shell script unit and smoke tests
.PHONY: test-scripts
test-scripts:
	@$(LINACTL) test-scripts

# Run Go unit tests for every workspace module with the race detector enabled.
# 对所有 Go workspace 模块运行单元测试，并启用竞态检测。
## test-go: Run Go unit tests with the race detector
.PHONY: test-go
test-go:
	@$(LINACTL) test-go $(TEST_GO_ARGS)

# Run only host-owned Playwright E2E tests. This target does not require the
# official plugin submodule.
## test-host: Run host-owned Playwright E2E tests without requiring official plugins
.PHONY: test-host
test-host:
	@$(LINACTL) test scope=host

# Run source-plugin-owned Playwright E2E tests. This target requires the
# official plugin submodule.
## test-plugins: Run official plugin Playwright E2E tests
.PHONY: test-plugins
test-plugins:
	@$(LINACTL) test scope=plugins
