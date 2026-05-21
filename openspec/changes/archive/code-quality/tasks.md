## 1. Explicit Dependency Injection

- [x] 1.1 扫描宿主和源码插件生产 Go 文件中的关键服务构造调用，按宿主 Controller、宿主 Service、Middleware、pkg/pluginservice/*、源码插件、WASM host service 分类输出清单
- [x] 1.2 更新 AGENTS.md 后端代码规范，新增显式依赖注入、禁止隐式关键服务构造、缓存敏感服务共享实例和测试豁免边界
- [x] 1.3 更新 lina-review 审查标准，新增后端依赖注入审查项
- [x] 1.4 增加静态扫描脚本或等价治理验证，识别生产路径中对关键服务 `New()` 的新增违规调用
- [x] 1.5 改造 auth.Service、middleware.Service 构造函数，显式接收依赖
- [x] 1.6 改造 role、menu、user、dict、file、usermsg、notify、sysconfig、i18n 等宿主服务构造入口
- [x] 1.7 改造 datascope、tenantcap、orgcap 等能力服务构造入口
- [x] 1.8 更新 cmd_http_runtime.go 现有 runtime 结构，使其持有共享的宿主服务实例
- [x] 1.9 改造宿主所有 Controller NewV1 构造函数，通过逐项参数接收服务依赖
- [x] 1.10 更新 cmd_http_routes.go，在路由绑定前从 runtime 共享实例构造 Controller
- [x] 1.11 为 pluginhost 的 HTTP/Cron registrar 新增宿主发布服务目录
- [x] 1.12 改造 pkg/pluginservice/* 适配器，使生产路径由宿主传入内部依赖
- [x] 1.13 迁移源码插件 backend/plugin.go 路由、全局中间件和 Cron 注册回调
- [x] 1.14 改造源码插件 Controller/Service 构造函数，移除隐式构造
- [x] 1.15 改造 WASM host service 配置入口，确保启动期注入共享实例
- [x] 1.16 为宿主路由构造和 WASM host service 添加单元测试
- [x] 1.17 运行宿主核心服务、Controller、源码插件后端单元测试
- [x] 1.18 运行新增静态扫描或治理验证
- [x] 1.19 修复 file.Service 与 jobmgmt.Service 构造函数内部创建 datascope.Service
- [x] 1.20 修复 WASM host service 配置入口在 nil 依赖时静默创建默认实例
- [x] 1.21 修复 plugin.New 在构造函数内部创建运行期关键依赖
- [x] 1.22 修复 plugin_startup_consistency.go 临时创建 tenantcap.Service 与 bizctx.Service
- [x] 1.23 修复 role.New 构造函数内部创建 datascope.Service
- [x] 1.24 修复 user.New 构造函数内部创建 datascope.Service
- [x] 1.25 收敛 pkg/pluginservice/* 契约到 contract 组件
- [x] 1.26 删除源码插件控制器 NewControllerV1 多余构造入口
- [x] 1.27 修复 pkg/sourceupgrade.New 隐式创建服务图
- [x] 1.28 修复 pluginservice/tenantfilter 通过包级全局配置读取宿主 bizctx
- [x] 1.29 将 ConfigureWasmHostServices 与 pluginhostservices.New 展开聚合依赖结构体
- [x] 1.30 将运行时初始化入口改为返回 error 而非 panic
- [x] 1.31 将源码插件注册与 registrar API 改为返回 error
- [x] 1.32 拆分 pluginhost_source_plugin.go 过长文件
- [x] 1.33 修复 multi-tenant 生命周期预检查通过 newTenantService(nil) 构造半初始化服务
- [x] 1.34 修复 panic allowlist 引用拆分前文件路径
- [x] 1.35 删除 role_new.go 中误追加的空 ControllerV1 和无参 NewV1

## 2. API Contract Hardening

- [x] 2.1 审查宿主 API 定义中直接嵌入或返回 entity.* 的响应类型
- [x] 2.2 为用户、文件、系统配置、字典、定时任务、任务日志和任务分组响应定义独立 DTO
- [x] 2.3 调整对应控制器响应映射，禁止直接把实体指针塞入 API 响应
- [x] 2.4 增加自动化测试或静态验证，覆盖 API 层不再依赖 internal/model/entity
- [x] 2.5 新增 pkg/listorder、pkg/tenantoverride、pkg/statusflag 公共契约组件
- [x] 2.6 调整 API 中重复的列表排序方向、租户覆盖模式、菜单类型、插件桥接类型和通用状态标志引用
- [x] 2.7 删除 API 包中对公共契约类型的兼容别名和常量转发
- [x] 2.8 统一源码插件 API DTO 命名和存放方式，移除 *Entity 命名
- [x] 2.9 迁移源码插件 API 契约单测到各插件自身测试目录
- [x] 2.10 清理文件详情页运行时翻译和宿主 apidoc i18n 中已不再暴露的响应字段翻译
- [x] 2.11 清理源码插件中 deletedAt 和旧 *Entity 响应 schema 的 apidoc 翻译键

## 3. Database Support Convergence

- [x] 3.1 从 pkg/dbdriver 移除 SQLite 驱动注册和支持类型
- [x] 3.2 从 pkg/dialect 移除 SQLite 方言实现、转译器、错误分类和相关测试
- [x] 3.3 调整依赖文件并运行 go mod tidy，移除 SQLite 驱动链路依赖残留
- [x] 3.4 更新后端测试，删除 SQLite 专属用例
- [x] 3.5 删除 SQLite smoke workflow、CI 输入和 main/nightly/release 调用参数
- [x] 3.6 删除 hack/tests 中 SQLite 专属 E2E 用例、support、脚本和 package scripts
- [x] 3.7 更新 linactl、makefile 注释和测试中 SQLite 配置示例
- [x] 3.8 更新运行时配置模板、packed 配置模板和镜像运行配置
- [x] 3.9 同步更新英文/中文 README 和测试文档中的数据库支持说明
- [x] 3.10 更新现行 OpenSpec 基线中 SQLite 支持口径
- [x] 3.11 修复 linapro-monitor-server PostgreSQL 单测未注册 pgsql 驱动
- [x] 3.12 修复 make init 在 clean checkout 中因 internal/packed/public 没有被跟踪的嵌入文件而无法编译
- [x] 3.13 将 plugindb/host 的元数据读与 schema probe 判定改为依赖 pkg/dialect 抽象

## 4. Verification

- [x] 4.1 运行宿主核心服务单元测试，覆盖 auth、session、middleware、role、datascope、config、i18n、plugin、cachecoord、kvcache、locker、cron
- [x] 4.2 运行宿主 Controller 和 cmd 路由相关单元测试
- [x] 4.3 运行所有源码插件后端单元测试
- [x] 4.4 运行 API/控制器编译烟测
- [x] 4.5 运行公共组件测试和 OpenSpec 校验
- [x] 4.6 运行至少覆盖变更包的 Go 编译/测试门禁
- [x] 4.7 运行工具链验证和静态扫描
- [x] 4.8 记录 i18n、缓存一致性、数据权限影响评估
- [x] 4.9 执行 lina-review 审查，修复发现的问题
