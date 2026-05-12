# LinaPro Test Commands
# LinaPro 测试指令
# ===================

# Run the complete Playwright E2E test suite.
# 运行完整的 Playwright E2E 测试套件。
## test: Run the full E2E test suite
.PHONY: test
test:
	@go run ./hack/tools/linactl test

# Run shell script unit and smoke tests for repository tooling.
# 运行仓库工具脚本的单元与 smoke 测试。
## test-scripts: Run shell script unit and smoke tests
.PHONY: test-scripts
test-scripts:
	@go run ./hack/tools/linactl test-scripts

# Run Go unit tests for every workspace module with the race detector enabled.
# 对所有 Go workspace 模块运行单元测试，并启用竞态检测。
## test-go: Run Go unit tests with the race detector
.PHONY: test-go
test-go:
	@go run ./hack/tools/linactl test-go
