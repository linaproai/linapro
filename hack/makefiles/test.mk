# LinaPro Test Commands
# LinaPro 测试指令
# ===================

# Run the complete Playwright E2E test suite.
# 运行完整的 Playwright E2E 测试套件。
## test: Run the full E2E test suite
.PHONY: test
test:
	@echo "Running E2E test suite..."
	cd hack/tests && pnpm test

# Run shell script unit and smoke tests for repository tooling.
# 运行仓库工具脚本的单元与 smoke 测试。
## test-scripts: Run shell script unit and smoke tests
.PHONY: test-scripts
test-scripts:
	@echo "Running shell script tests..."
	@for test_file in hack/tests/scripts/*.sh; do \
		echo "==> $$test_file"; \
		bash "$$test_file"; \
	done
