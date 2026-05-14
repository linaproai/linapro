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

# Run cross-platform repository tool smoke checks.
# 运行跨平台仓库工具 smoke 检查。
## test.scripts: Run cross-platform repository tool smoke checks
.PHONY: test.scripts
test.scripts:
	@$(LINACTL) test.scripts

# Run Go unit tests for every workspace module with the race detector enabled.
# 对所有 Go workspace 模块运行单元测试，并启用竞态检测。
## test.go: Run Go unit tests with the race detector
.PHONY: test.go
test.go:
	@$(LINACTL) test.go $(TEST_GO_ARGS)

# Run go mod tidy in every maintained Go module directory.
# 在每个受维护的 Go 模块目录下运行 go mod tidy。
## tidy: Run go mod tidy in every Go module
.PHONY: tidy
tidy:
	@$(LINACTL) tidy

# Run only host-owned Playwright E2E tests. This target does not require the
# official plugin submodule.
## test.host: Run host-owned Playwright E2E tests without requiring official plugins
.PHONY: test.host
test.host:
	@$(LINACTL) test.host

# Run source-plugin-owned Playwright E2E tests. This target requires the
# official plugin submodule.
## test.plugins: Run official plugin Playwright E2E tests
.PHONY: test.plugins
test.plugins:
	@$(LINACTL) test.plugins
