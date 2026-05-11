# LinaPro 插件目录

`apps/lina-plugins/` 是 LinaPro 官方源码插件工作区。

在当前开源阶段，宿主只保留稳定的核心基础能力，例如用户管理、角色管理、菜单管理、字典管理、参数设置、文件管理、任务调度、插件治理和开发支持；非核心业务能力通过 `apps/lina-plugins/<plugin-id>/` 下的源码插件交付。

## 这里包含什么

当前目录下包含以下参考内容：

- `plugin-demo-source`：源码插件目录结构与开发方式样例
- `plugin-demo-dynamic`：动态 WASM 插件结构与生命周期样例
- 官方源码插件：通过显式接线编译进宿主的一方业务插件

## 官方源码插件列表

当前仓库内置以下一方源码插件：

- `demo-control`：演示环境只读保护
- `org-center`：部门管理、岗位管理
- `content-notice`：通知公告管理
- `monitor-online`：在线用户查询与强制下线治理
- `monitor-server`：服务监控采集、清理与查询
- `monitor-operlog`：操作日志落库与治理
- `monitor-loginlog`：登录日志落库与治理

每个官方插件都使用统一的基础结构：

```text
apps/lina-plugins/<plugin-id>/
  backend/
    api/                插件 API DTO 与路由契约
    internal/
      controller/       插件 HTTP 控制器
      service/          插件业务服务
      dao/              插件存在数据库访问时生成的本地 DAO 工件
      model/do/         插件存在数据库访问时生成的本地 DO 工件
      model/entity/     插件存在数据库访问时生成的本地实体工件
    hack/config.yaml    插件本地 GoFrame codegen 配置
    plugin.go           插件后端注册入口
  frontend/pages/       由宿主菜单挂载的插件页面
  manifest/sql/         插件自有安装 SQL 资源
  manifest/sql/mock-data/ 插件自有可选`mock`/演示 SQL 资源
  manifest/sql/uninstall/ 插件自有卸载 SQL 资源
  hack/tests/e2e/       可选的插件自有 E2E TC 用例
  hack/tests/pages/     可选的插件自有 E2E 页面对象
  hack/tests/support/   可选的插件自有 E2E helper
  plugin.yaml           插件清单
  plugin_embed.go       嵌入资源注册入口
  README.md             英文说明
  README.zh-CN.md       中文说明
```

`backend/internal/service/` 是源码插件业务 `service` 的唯一合法目录，禁止再创建 `backend/service/`。

## 宿主与插件边界

当前源码插件方案强调通过稳定接缝解耦，而不是在宿主里散落大量 `if pluginEnabled` 判断。

- 宿主拥有稳定的一级目录骨架，例如 `dashboard`、`iam`、`setting`、`scheduler`、`extension`、`developer`。
- 插件菜单只能挂载到宿主已发布的稳定目录，或者插件自身声明的内部菜单节点。
- 官方插件挂载点固定：`org-center -> org`、`content-notice -> content`、全部监控插件 -> `monitor`。
- 宿主对插件发布稳定能力接缝，例如认证事件、审计事件、组织能力接口和插件生命周期 Hook。
- 插件自有的数据表、菜单、页面、Hook 和定时任务都保留在插件目录内，并通过插件生命周期完成安装与卸载。

## 源码插件开发流程

1. 创建 `apps/lina-plugins/<plugin-id>/`。
2. 参考 `plugin-demo-source/` 的目录结构。
3. 在 `plugin.yaml` 中声明清单、菜单、页面、SQL 资源与可选 Hook。
4. 插件后端代码保留在插件目录中，业务逻辑统一放在 `backend/internal/service/` 下，并且只依赖宿主公开包。
5. 在 `apps/lina-plugins/lina-plugins.go` 中做显式接线。

## 插件自有 E2E 测试

源码插件应把插件专属 Playwright 覆盖放在 `apps/lina-plugins/<plugin-id>/hack/tests/e2e/` 下。
插件页面对象和辅助 helper 应分别保留在同级的 `hack/tests/pages/` 与 `hack/tests/support/` 中。
宿主测试运行器会通过通用 `plugins` 范围发现这些测试；单个插件也可以通过 `pnpm -C hack/tests test:module -- plugin:<plugin-id>` 直接运行，不需要为每个插件在执行清单里新增专属 scope。

## 源码插件版本升级

当某个源码插件已经在宿主中安装完成，而你又提升了它的 `plugin.yaml` 版本时，源码扫描不再自动替换当前生效版本。

- 当前生效版本仍然固定在 `sys_plugin.version` 与 `release_id`。
- 新发现的更高源码版本会写入一个 `prepared` 状态的 `sys_plugin_release`。
- 在宿主允许启动前，必须通过 AI 工具调用 `lina-upgrade` 技能，例如 `upgrade source plugin <plugin-id>` 或 `upgrade all source plugins`。
- 如果跳过这一步，宿主启动会直接失败，并输出需要调用的技能提示。

## 动态插件说明

动态 WASM 插件仍然适用于运行时托管交付场景。如果插件需要通过上传、安装、启用、停用和卸载完成生命周期管理，请参考 `plugin-demo-dynamic/`。

## 参考入口

- `apps/lina-plugins/plugin-demo-source/README.md`
- `apps/lina-plugins/plugin-demo-dynamic/README.md`
- `apps/lina-plugins/OPERATIONS.md`
