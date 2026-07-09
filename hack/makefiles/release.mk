# LinaPro Release Governance Commands
# LinaPro 发布治理指令
# ===============================

RELEASE_ARGS :=
VERSION_ARGS :=
ifneq ($(origin tag), undefined)
RELEASE_ARGS += tag=$(tag)
endif
ifneq ($(origin metadata), undefined)
RELEASE_ARGS += metadata=$(metadata)
endif
ifneq ($(origin print-version), undefined)
RELEASE_ARGS += print-version=$(print-version)
endif
ifneq ($(origin to), undefined)
VERSION_ARGS += to=$(to)
endif

# Verify that the release tag matches framework.version in metadata.yaml.
# 校验 release tag 与 metadata.yaml 中的 framework.version 一致。
## release.tag.check: Verify release tag equals apps/lina-core/manifest/config/metadata.yaml framework.version
.PHONY: release.tag.check
release.tag.check:
	@$(LINACTL) release.tag.check $(RELEASE_ARGS)

# Update framework metadata version and README image cache query parameters.
# 更新框架元数据版本号，并刷新 README 图片缓存查询参数。
## version: Update apps/lina-core/manifest/config/metadata.yaml framework.version and README image cache keys; use to=v0.1.0
.PHONY: version
version:
	@$(LINACTL) version $(VERSION_ARGS)
