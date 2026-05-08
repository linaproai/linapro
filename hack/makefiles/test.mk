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

# Run Go unit tests for every workspace module with the race detector enabled.
# 对所有 Go workspace 模块运行单元测试，并启用竞态检测。
## test-go: Run Go unit tests with the race detector
.PHONY: test-go
test-go:
	@echo "Running Go unit tests with race detector..."
	@set -e; \
	module_file="$$(mktemp)"; \
	go list -m -f '{{.Dir}}' > "$$module_file"; \
	if [ ! -s "$$module_file" ]; then \
		echo "No Go workspace modules discovered"; \
		exit 1; \
	fi; \
	while IFS= read -r module_dir; do \
		if [ -z "$$module_dir" ]; then \
			continue; \
		fi; \
		echo "==> go test -race -v $$module_dir"; \
		(cd "$$module_dir" && go test -race -v ./...); \
	done < "$$module_file"
