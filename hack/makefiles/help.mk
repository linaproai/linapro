# LinaPro Help Commands
# LinaPro 帮助指令
# =================

# Print the available root Make targets from this file and included target files.
# 打印根 Makefile 及其引入目标文件中可用的 make 目标。
## help: Show help
.PHONY: help
help:
	@go run ./hack/tools/linactl help
