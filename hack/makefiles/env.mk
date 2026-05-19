# LinaPro Environment Commands
# LinaPro 环境检查与初始化指令
# ===========================

# Check local development tool versions without modifying the workspace.
# 检查本地开发工具版本，不修改工作区。
## env.check: Check local development tool versions
.PHONY: env.check
env.check:
	@$(LINACTL) env.check

# Install all prerequisites for development and E2E tests (frontend deps + Playwright browsers).
# 安装开发与 E2E 测试的所有前置依赖（前端依赖 + Playwright 浏览器）。
## env.setup: Install frontend dependencies and Playwright browsers
.PHONY: env.setup
env.setup:
	@$(LINACTL) env.setup
