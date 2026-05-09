# LinaPro Help Commands
# LinaPro 帮助指令
# =================

# Print the available root Make targets from this file and included target files.
# 打印根 Makefile 及其引入目标文件中可用的 make 目标。
## help: Show help
.PHONY: help
help:
	@set -e; \
	if [ -t 1 ]; then \
		c_title='\033[1;36m'; \
		c_cmd='\033[1;32m'; \
		c_dim='\033[2m'; \
		c_reset='\033[0m'; \
	else \
		c_title=''; \
		c_cmd=''; \
		c_dim=''; \
		c_reset=''; \
	fi; \
	printf "$${c_dim}Usage:$${c_reset} make $${c_cmd}<target>$${c_reset}\n\n"; \
	awk '/^## [^:]+:/ { \
		line=$$0; \
		sub(/^## /, "", line); \
		sep=index(line, ": "); \
		if (sep > 0) { \
			name=substr(line, 1, sep - 1); \
			desc=substr(line, sep + 2); \
			printf "%s\t%s\n", name, desc; \
		} \
	}' $(MAKEFILE_LIST) | sort -k1,1 | \
	awk -F '\t' -v c_cmd="$$c_cmd" -v c_dim="$$c_dim" -v c_reset="$$c_reset" ' \
		{ \
			names[++count]=$$1; \
			descs[count]=$$2; \
			if (length($$1) > max) { \
				max=length($$1); \
			} \
		} \
		END { \
			print c_dim "Available targets:" c_reset; \
			for (i=1; i<=count; i++) { \
				printf "  %s%-*s%s  %s\n", c_cmd, max, names[i], c_reset, descs[i]; \
			} \
		}'
