# LinaPro I18n Targets
# ===================

## check-runtime-i18n: 扫描运行时可见硬编码文案
.PHONY: check-runtime-i18n
check-runtime-i18n:
	@go run ./hack/tools/runtime-i18n scan

## check-runtime-i18n-messages: 校验宿主与插件运行时语言包 key 覆盖
.PHONY: check-runtime-i18n-messages
check-runtime-i18n-messages:
	@go run ./hack/tools/runtime-i18n messages
