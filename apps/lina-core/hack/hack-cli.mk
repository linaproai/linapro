
# Install/update to the latest GoFrame CLI tool.
# 安装或更新到最新的 GoFrame CLI 工具。
.PHONY: cli
cli:
	@go run ../../hack/tools/linactl cli


# Check and install the GoFrame CLI tool when missing.
# 检查 GoFrame CLI 工具，缺失时自动安装。
.PHONY: cli.install
cli.install:
	@go run ../../hack/tools/linactl cli.install
