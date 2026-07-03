## Why

插件`distribution`字段当前使用`marketplace`表示普通可管理插件，但该命名容易误导为插件必须来自线上市场或商业分发渠道。实际语义是插件仍由插件管理入口或`plugin.autoEnable`显式治理。为了让插件生命周期治理契约更准确、稳定，需要将普通分发治理枚举重命名为`managed`，并保留`builtin`表示随宿主编译交付的内建源码插件。

## What Changes

- 将`distribution`合法值从`marketplace | builtin`调整为`managed | builtin`，省略字段时默认归一化为`managed`。
- 更新宿主`API`枚举、manifest 校验、release snapshot 校验、数据库默认值、插件清单、前端类型、文档和测试期望。
- 不保留`marketplace`兼容分支；当前项目无历史包袱，旧值在有效契约中应视为非法值。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `plugin-manifest-lifecycle`：更新插件 manifest 分发治理枚举，将普通可管理插件从`marketplace`改为`managed`。

## Impact

- 影响`apps/lina-core`插件公开`API`、插件 manifest 校验、SQL 基线、发布快照校验、插件清单、前端插件类型和相关说明文档。
- `i18n`影响判断：修改宿主`API`文档源文本、中文`apidoc`翻译资源和插件清单说明；不新增运行时 UI 文案、菜单或路由。
- 缓存一致性影响判断：仅修改枚举值与静态清单/文档，不新增缓存、失效、预热或跨实例同步机制。
- 数据权限影响判断：不新增或修改读取、写入、导出、下载、聚合统计、批量信息或执行类接口的数据权限边界。
- 开发工具跨平台影响判断：不新增或修改`Makefile`、脚本、`linactl`命令、`CI`或构建工具入口。
