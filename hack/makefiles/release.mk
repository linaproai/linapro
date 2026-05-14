# LinaPro Release Governance Commands
# LinaPro 发布治理指令
# ===============================

RELEASE_ARGS :=
ifneq ($(origin tag), undefined)
RELEASE_ARGS += tag=$(tag)
endif
ifneq ($(origin metadata), undefined)
RELEASE_ARGS += metadata=$(metadata)
endif
ifneq ($(origin print_version), undefined)
RELEASE_ARGS += print_version=$(print_version)
endif
ifneq ($(origin print-version), undefined)
RELEASE_ARGS += print-version=$(print-version)
endif

# Verify that the release tag matches framework.version in metadata.yaml.
# 校验 release tag 与 metadata.yaml 中的 framework.version 一致。
## release.tag.check: Verify release tag equals apps/lina-core/manifest/config/metadata.yaml framework.version
.PHONY: release.tag.check
release.tag.check:
	@$(LINACTL) release.tag.check $(RELEASE_ARGS)
