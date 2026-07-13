# 数据库与 SQL 规则

## 适用范围

本规则约束数据库迁移 SQL、Seed DML、Mock 数据、插件安装卸载 SQL、DAO 生成、数据库幂等执行、自增主键写入策略、索引设计和查询性能基础。

## 数据表设计要求

在设计新的业务模块时，无论是主框架还是业务插件的业务模块，都应当和用户确认以下数据表设计点：

- **业务模块是否需要使用软删除设计**：以便确定对应的业务模块数据表是否需要增加`deleted_at`字段，编码时是否需要使用到软删除功能。

## 插件业务表命名要求

- 插件在`manifest/sql/`中新增或维护的**插件自有业务表**，表名必须使用**完整插件 ID** 作为前缀，不得使用截断或臆造的短前缀。
- 命名格式：`plugin_<plugin_id_snake>_<entity>`
  - `<plugin_id_snake>`：将`plugin.yaml`中的`id`（kebab-case）中的`-`全部替换为`_`，**保留完整 ID 段**（例如`linapro-mail-core` → `linapro_mail_core`，不得写成`linapro_mail`）。
  - `<entity>`：表业务实体名，小写 snake_case（例如`connection`、`account`、`dept`）。
- 索引、唯一约束等对象名称应与表前缀保持一致，便于按插件识别与卸载。
- 正确示例：
  - 插件`linapro-org-core` → `plugin_linapro_org_core_dept`
  - 插件`linapro-tenant-core` → `plugin_linapro_tenant_core_tenant`
  - 插件`linapro-mail-core` → `plugin_linapro_mail_core_connection`
  - 插件`linapro-demo-source` → `plugin_linapro_demo_source_record`
- 错误示例：
  - 插件`linapro-mail-core`却使用`plugin_linapro_mail_connection`（缺少 ID 中的`core`段）
- 宿主自有表继续使用既有`sys_*`等宿主命名约定，不适用本条插件前缀规则。
- 插件`make dao`的`tables`与`removePrefix`必须与上述完整前缀一致（例如`removePrefix: "plugin_linapro_mail_core_"`）。

## Go 代码生成流程

- 数据库变更应新增或修改当前迭代对应的`manifest/sql/{序号}-{迭代名称}.sql`，执行`make db.init`更新数据库，再执行`make dao`生成或更新 Go 源码文件。
- `DAO/DO/Entity`源码文件由`make dao`自动生成，不要手动创建或修改。

## SQL 文件命名和版本管理

- 数据库变更 SQL 文件采用`{序号}-{当前迭代名称}.sql`格式命名，存放在`manifest/sql/`目录下。
- 序号为三位数字，例如`001`、`002`。
- 当前迭代不涉及数据库变更时，不生成该迭代的 SQL 文件。
- 每次迭代应新建 SQL 文件维护数据库变更，而非修改旧迭代创建的 SQL 文件。
- 仅在用户明确要求时才允许修改旧迭代 SQL 文件。
- 宿主`manifest/sql/`目录下，同一个业务迭代只保留一个版本 SQL 文件。
- 当前迭代后续继续发生数据库变更时，应追加或整理到当前迭代对应的同一个 SQL 文件中，而不是再新增第二个同迭代 SQL 文件。

## SQL 数据分类要求

- 迭代 SQL 文件中只允许包含 DDL 和 Seed DML。
- Seed DML 指系统运行所必需的初始化数据，例如字典类型、管理员账号等。
- 演示或测试用 Mock 数据必须放到`manifest/sql/mock-data/`目录下的独立 SQL 文件中。
- Mock 数据文件名以数字前缀控制执行顺序，例如`01_mock_depts.sql`、`02_mock_posts.sql`。
- 插件安装 SQL 放在插件自己的`manifest/sql/`，卸载 SQL 放在`manifest/sql/uninstall/`。

## SQL 幂等性要求

- 所有交付到`manifest/sql/`、`manifest/sql/mock-data/`以及插件`manifest/sql/`的建表、改表、索引变更和数据写入脚本都必须满足可重复执行且结果一致的幂等性要求。
- 当前项目默认源 SQL 方言为 PostgreSQL。
- 编写 SQL 时应优先使用带存在性保护的语法，例如`CREATE TABLE IF NOT EXISTS`、`CREATE INDEX IF NOT EXISTS`、`DROP ... IF EXISTS`、`ALTER TABLE ... ADD COLUMN IF NOT EXISTS`、`INSERT ... ON CONFLICT DO NOTHING`。
- 若目标语句或数据库版本不支持直接的`IF [NOT] EXISTS`或`ON CONFLICT`语法，必须通过前置存在性判断或等价安全重入方案实现幂等。
- 禁止提交只能成功执行一次的 SQL 脚本。

## 数据写入要求

- 交付到宿主或插件`manifest/sql/`的 Seed DML、初始化数据和 Mock 数据脚本，统一只允许使用`INSERT ... ON CONFLICT DO NOTHING`或前置存在性判断实现幂等。
- `ON CONFLICT DO NOTHING`必须依赖目标表真实业务语义下的`PRIMARY KEY`、`UNIQUE`约束或唯一索引。
- 禁止在没有实际冲突依据的表上机械使用`ON CONFLICT DO NOTHING`。
- 日志、历史、监控流水等不应为 Mock 幂等而新增会限制真实业务写入的唯一约束。
- 少量静态演示行应使用覆盖演示身份字段的精确`NOT EXISTS`判断保持重复加载结果一致。
- 禁止通过`INSERT INTO ... ON DUPLICATE KEY UPDATE`在重复执行时更新已有记录。
- 向自增主键表写入 Seed DML、Mock 数据或插件安装初始化数据时，禁止在`INSERT`列清单中显式指定自增`id`值。
- 需要建立关联关系时，应通过稳定业务键、唯一键或子查询解析目标记录，而不是硬编码自增`id`。

## 查询性能与索引要求

- 新增或修改数据表、字段、关联关系和查询语义时，必须同步评估列表筛选、排序、分页、聚合、树形查询、关联装配、租户隔离、组织过滤、数据权限过滤和软删除过滤所需的索引。
- 常用查询路径需要依赖索引、唯一约束、稳定业务键或可批量关联的外键完成，禁止让核心列表、下拉、导出、聚合或权限过滤默认依赖全表扫描。
- 索引设计必须基于真实查询路径和字段选择性，不得为了规避审查机械添加低价值索引；组合索引的字段顺序必须匹配主要过滤、排序或关联方式。
- 查询设计必须优先支持集合化读取、批量关联和数据库侧聚合，避免只能由应用层循环逐项查询才能完成的数据模型。
- 不得在高频过滤或排序条件中依赖会破坏普通索引使用的字段函数转换；确需派生值时，应在模型或迁移设计中明确可查询字段、索引或缓存方案。
- 新增索引、唯一约束或查询辅助字段必须满足 SQL 幂等性要求，并与当前迭代 SQL 文件一起维护。

## 验证要求

- SQL 变更必须运行匹配的数据库初始化、迁移、生成或静态检查入口。
- 至少必须执行`openspec validate <change> --strict`和 SQL 静态检索或审查。
- DAO 生成变更必须运行覆盖生成结果使用包的 Go 编译门禁。
- SQL 幂等性、数据分类、自增主键写入和软删除语义必须在任务记录或审查结论中明确验证。
- 涉及高频查询路径时，必须在任务记录或审查结论中说明索引、集合化查询、批量关联或缓存方案如何避免`N+1`查询和无约束全表扫描。

## 审查要求

- 审查必须确认 SQL 文件命名、版本管理和单迭代单文件规则。
- 审查必须确认 Seed DML 与 Mock 数据分离。
- 审查必须确认 SQL 可安全重复执行。
- 审查必须拒绝`ON DUPLICATE KEY UPDATE`和显式写入自增`id`。
- 审查必须确认软删除和自动时间维护没有被手写冗余逻辑破坏。
- 审查必须确认高频查询路径具备合理索引、批量关联或缓存方案，拒绝只能通过应用层循环查询完成的数据模型。
