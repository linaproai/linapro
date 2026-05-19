# 设计

## 设计原则

本次采用 Kubernetes 风格的契约归属策略：

1. 跨 API 的基础契约进入稳定公共包。
2. 领域私有枚举留在所属领域，不为消除重复而强行上提。
3. API DTO 字段直接引用稳定公共包类型，不在 API 包内为公共契约创建兼容别名或常量转发。
4. 公共组件保持小而明确，禁止新增大一统 `pkg/enums`。

## 公共组件划分

### `pkg/listorder`

承载列表查询排序方向：

- `Direction`
- `ASC`
- `DESC`
- `Normalize`
- `IsSupported`

API 层列表请求直接使用 `listorder.Direction`，控制器在进入服务层前按字符串语义转换。

### `pkg/tenantoverride`

承载平台默认数据和租户覆盖数据之间的操作模式：

- `Mode`
- `None`
- `CreateTenantOverride`
- `Normalize`
- `CanCreateTenantOverride`

配置管理和字典管理共享该契约。

### `pkg/statusflag`

承载常见 `0/1` 状态标志的值类型与常量：

- `Enabled` / `Disabled`
- `Visible` / `Hidden`
- `Installed` / `Uninstalled`
- `Read` / `Unread`
- `Yes` / `No`

该组件只表达通用标志，不定义万能 `Status`，避免把业务状态、安装状态、可见性和已读语义混为一谈。

### 复用已有组件

- 菜单类型复用 `pkg/menutype`，API DTO 字段直接使用 `menutype.Code`。
- 插件动态路由访问模式和运行时类型复用 `pkg/pluginbridge` 常量，API DTO 字段直接使用字符串承载桥接契约值，控制器转换点直接引用 `pluginbridge`。
- 定时任务领域枚举仍保留在 API 包和 `internal/service/jobmeta` 各自边界内。本次仅抽象各列表 API 共同使用的排序方向；如后续需要让源码插件或公开组件直接引用定时任务领域契约，再单独评估是否提升为 `pkg/jobmeta`。

## API 兼容性

所有外部 JSON 字段名、取值和默认值保持不变。Go 类型变更仅影响内部编译期引用：

- 排序方向、租户覆盖模式、菜单类型和通用状态标志在 DTO 中直接暴露为所属公共包类型。
- 插件桥接常量值仍为 `wasm`、`public`、`login`，但 API 包不再维护二次常量转发。

## i18n 影响

本次不新增、修改或删除用户可见文本、菜单、按钮、表单、表格列、提示信息或运行时翻译资源。API 文档 `dc` 和 `eg` 取值保持英文源文本和原示例语义，不新增 apidoc i18n 资源。

## 缓存与分布式一致性影响

本次仅抽象 Go 编译期类型和值常量，不新增缓存、不修改缓存键、不改变运行时缓存失效路径，也不影响分布式部署一致性。

## 数据权限影响

本次不新增或修改数据操作接口，不改变查询、详情、更新、删除或导出行为；数据权限边界保持现状。

## 验证策略

- 新增公共组件单元测试覆盖 `Normalize`、`String`、`Int` 和支持性判断。
- 运行 `cd apps/lina-core && go test ./pkg/listorder ./pkg/tenantoverride ./pkg/statusflag -count=1`。
- 运行覆盖 API DTO 和控制器引用的编译烟测：`cd apps/lina-core && go test ./api ./internal/controller/... -count=1`。
- 如果控制器构造或路由绑定受到 API 类型变更影响，运行 `cd apps/lina-core && go test ./internal/cmd -count=1`。
