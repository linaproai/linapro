# LinaPro Database Commands
# LinaPro 数据库指令
# =====================

# Initialize the backend database with schema and required seed data.
# The backend dispatches by database.default.link, for example PostgreSQL or SQLite.
# 初始化后端数据库表结构和系统必需的种子数据。
# 后端会按 database.default.link 自动分发到 PostgreSQL 或 SQLite 等方言。
## init: Initialize the database with DDL and seed data only
.PHONY: init
init:
	@if [ "$(confirm)" != "init" ]; then \
		echo "✗ make init requires explicit confirmation for safety"; \
		echo "  Use: make init confirm=init"; \
		echo "  To rebuild the linapro database: make init confirm=init rebuild=true"; \
		exit 1; \
	fi
	@cd $(BACKEND_DIR) && $(MAKE) init confirm=$(confirm) $(if $(rebuild),rebuild=$(rebuild),)

# Load optional mock data for local demos and development verification.
# Mock loading uses the same database.default.link dialect and requires init first.
# 加载用于本地演示和开发验证的可选 Mock 数据。
# Mock 加载使用同一个 database.default.link 方言，并要求先完成 init。
## mock: Load mock demo data after init
.PHONY: mock
mock:
	@if [ "$(confirm)" != "mock" ]; then \
		echo "✗ make mock requires explicit confirmation for safety"; \
		echo "  Use: make mock confirm=mock"; \
		exit 1; \
	fi
	@cd $(BACKEND_DIR) && $(MAKE) mock confirm=$(confirm)
