## Why

默认管理工作台存在两组影响日常使用的问题：

1. 部门管理页面在部门树为空时无法创建首个顶级部门。后端 `POST /dept` 已支持 `parentId=0` 创建顶级部门，但前端新增抽屉将上级部门设为必选且候选项完全来自现有部门树，导致空树场景下没有可选项，也不会提交合法的 `parentId=0`。

2. 字典标签更新后已打开列表未同步显示最新标签。字典 Store 的 `resetCache()` 直接清空内部 Map，但已挂载组件持有的是此前返回的数组引用，清空 Map 不会更新这些旧数组。

3. 切换工作台 Tab 后前一个列表页分页状态丢失。内建列表菜单 seed 的 `is_cache=0`，导致 `/menus/all` 投影 `meta.keepAlive=false`，前端不缓存这些页面实例。

4. 动态插件样例的 CI 单测因 `.gitignore` 忽略的 `config.yaml` 缺失而失败，且此前修复误将运行配置契约从 `config.yaml` 改为 `config.example.yaml`，需纠正并让单测自行准备 fixture。

## What Changes

- 修复 `linapro-org-core` 插件部门新增抽屉：在上级部门 TreeSelect 中合成本地"顶级部门"节点，空树和普通新增场景默认提交 `parentId=0`，候选数据统一投影为 `id`、`label`、`children` 结构。
- 修复字典 Store：支持按字典类型精确刷新，对原数组执行就地 splice 保持引用稳定；字典数据新增、更新、删除后按当前 dictType 刷新，字典类型 key 变更时同步刷新新旧 key。
- 修正宿主菜单 Seed DML：将内建可分页 routed 菜单的 `is_cache` 设置为 `1`，直接修正源头 seed 而非新增补丁 SQL。
- 纠正动态插件样例配置契约：保持 `config/config.yaml` 作为运行期配置，单测在 CI 环境从 `config.example.yaml` 准备临时 fixture。

## Capabilities

| 能力 | 说明 |
|------|------|
| 部门空树创建 | 部门树为空时可直接创建首个顶级部门 |
| 字典标签同步 | 字典数据变更后已打开列表标签实时刷新 |
| Tab 分页保持 | 工作台 Tab 切换后列表页保留页码和分页大小 |
| 动态插件 CI 稳定 | 动态插件样例单测在 CI 干净 checkout 中通过 |

## Impact

- 宿主边界：仅修改 `linapro-org-core` 插件前端适配层和宿主菜单 Seed DML，不修改 `apps/lina-core` 核心领域契约、通用 service 语义或存储模型。
- API 契约：不新增或修改 HTTP API、路由、DTO、权限标签或响应结构。
- 数据权限：不新增数据访问路径；创建动作继续走现有后端租户过滤和权限校验。
- 缓存一致性：涉及前端进程内字典缓存和工作台 Tab keep-alive，不涉及后端分布式缓存。
- 数据库：修改宿主 Seed DML 中内建菜单 `is_cache` 初始值，不修改表结构、索引或 DAO。
- 后端 Go：仅修改 `linactl` Go 单测 fixture 准备逻辑，无生产 Go 代码变更。
- i18n：新增插件运行时文案"顶级部门"，维护插件自身 `zh-CN` 与 `en-US` 资源。
- 测试：新增部门空树创建 E2E、字典标签同步与 Tab 分页保持 E2E、字典 Store 单元测试。
