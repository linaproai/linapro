# LinaPro I18n Commands
# LinaPro 国际化检查指令
# ===================

# Run runtime-visible hard-coded text scanning and message key coverage checks.
# 运行运行时可见硬编码文案扫描和消息 key 覆盖校验。
## i18n.check: Run runtime i18n governance checks
.PHONY: i18n.check
i18n.check:
	@$(LINACTL) i18n.check
