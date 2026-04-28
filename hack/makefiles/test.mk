# LinaPro Test Targets
# ===================

## test: 运行完整 E2E 测试套件
.PHONY: test
test:
	@echo "🧪 运行 E2E 测试套件..."
	cd hack/tests && pnpm test

## test-install: 运行安装脚本 smoke test（适用于本地与 CI）
.PHONY: test-install
test-install:
	@echo "🧪 运行安装脚本 smoke test..."
	@python3 hack/scripts/install/test_install.py
