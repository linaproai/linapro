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

# Run installer smoke tests for local development and CI verification.
# 运行适用于本地开发和 CI 校验的安装脚本 smoke test。
## test-install: Run installer smoke tests for local and CI verification
.PHONY: test-install
test-install:
	@echo "Running installer smoke tests..."
	@bash hack/tests/scripts/install-bootstrap.sh all
