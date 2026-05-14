# LinaPro Framework - Root Makefile
# ===========================

BACKEND_DIR   := apps/lina-core
FRONTEND_DIR  := apps/lina-vben
TEMP_DIR      := temp
PID_DIR       := $(TEMP_DIR)/pids
BACKEND_PID   := $(PID_DIR)/backend.pid
FRONTEND_PID  := $(PID_DIR)/frontend.pid
BACKEND_PORT  := 8080
FRONTEND_PORT := 5666
BACKEND_LOG   := $(TEMP_DIR)/lina-core.log
FRONTEND_LOG  := $(TEMP_DIR)/lina-vben.log
EMBED_DIR     := $(BACKEND_DIR)/internal/packed/public
OUTPUT_DIR    := $(TEMP_DIR)/output
LINACTL       := cd hack/tools/linactl && go run .

# Include split makefile targets.
# 引入拆分后的 Makefile 目标文件。
include hack/makefiles/dev.mk
include hack/makefiles/build.mk
include hack/makefiles/image.mk
include hack/makefiles/test.mk
include hack/makefiles/i18n.mk
include hack/makefiles/database.mk
include hack/makefiles/help.mk
