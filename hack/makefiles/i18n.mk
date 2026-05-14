# LinaPro I18n Commands
# LinaPro 国际化检查指令
# ===================

# Scan runtime-visible code paths for hard-coded text.
# 扫描运行时可见代码路径中的硬编码文案。
## check-runtime-i18n: Scan runtime-visible hard-coded text
.PHONY: check-runtime-i18n
check-runtime-i18n:
	@$(LINACTL) check-runtime-i18n

# Validate runtime i18n message key coverage for host and plugin resources.
# 校验宿主和插件运行时语言包的消息 key 覆盖情况。
## check-runtime-i18n-messages: Validate host and plugin runtime i18n message key coverage
.PHONY: check-runtime-i18n-messages
check-runtime-i18n-messages:
	@$(LINACTL) check-runtime-i18n-messages
