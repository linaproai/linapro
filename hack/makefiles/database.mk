# LinaPro Database Commands
# LinaPro 数据库指令
# =====================

# Initialize the backend database with schema and required seed data.
# The backend dispatches by database.default.link and currently supports PostgreSQL.
# 初始化后端数据库表结构和系统必需的种子数据。
# 后端会按 database.default.link 自动分发，目前仅支持 PostgreSQL 方言。
## db.init: Initialize the database with DDL and seed data only
.PHONY: db.init
db.init:
	@$(LINACTL) db.init confirm=$(confirm) $(if $(rebuild),rebuild=$(rebuild),)

# Replay all host framework SQL files for database upgrades.
# The SQL files are idempotent, so this upgrades existing tables and seed data
# to the latest delivered state without rebuilding the configured database.
# 重放主框架全部 SQL 文件用于数据库升级。
# SQL 文件是幂等的，因此会将已有表结构和种子数据升级到最新交付状态，
# 不会重建配置中的数据库。
## db.upgrade: Replay host SQL files to upgrade the database
.PHONY: db.upgrade
db.upgrade:
	@$(LINACTL) db.upgrade confirm=$(confirm)

# Load optional mock data for local demos and development verification.
# Mock loading uses the same database.default.link dialect and requires db.init first.
# 加载用于本地演示和开发验证的可选 Mock 数据。
# Mock 加载使用同一个 database.default.link 方言，并要求先完成 db.init。
## db.mock: Load mock demo data after db.init
.PHONY: db.mock
db.mock:
	@$(LINACTL) db.mock confirm=$(confirm)

# Generate GoFrame controller scaffolding. Defaults to the host backend; pass
# dir=<backend-dir> to target another backend explicitly.
# 生成 GoFrame 控制器骨架。默认面向宿主后端；通过 dir=<backend-dir>
# 显式指定其他后端目标。
## ctrl: Generate GoFrame controller scaffolding
.PHONY: ctrl
ctrl:
	@$(LINACTL) ctrl $(if $(dir),dir="$(dir)",)

# Generate DAO/DO/Entity files. Defaults to the host backend; pass
# dir=<backend-dir> to target another backend explicitly.
# 生成 DAO/DO/Entity 文件。默认面向宿主后端；通过 dir=<backend-dir>
# 显式指定其他后端目标。
## dao: Generate GoFrame DAO/DO/Entity files
.PHONY: dao
dao:
	@$(LINACTL) dao $(if $(dir),dir="$(dir)",)
