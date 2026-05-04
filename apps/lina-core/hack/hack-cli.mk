
# Install/update to the latest GoFrame CLI tool.
# 安装或更新到最新的 GoFrame CLI 工具。
.PHONY: cli
cli:
	@set -e; \
	wget -O gf \
	https://github.com/gogf/gf/releases/latest/download/gf_$(shell go env GOOS)_$(shell go env GOARCH) && \
	chmod +x gf && \
	./gf install -y && \
	rm ./gf


# Check and install the GoFrame CLI tool when missing.
# 检查 GoFrame CLI 工具，缺失时自动安装。
.PHONY: cli.install
cli.install:
	@set -e; \
	gf -v > /dev/null 2>&1 || if [[ "$?" -ne "0" ]]; then \
		echo "GoFrame CLI is not installed; starting automatic installation..."; \
		make cli; \
	fi;
