## Overview

本改进在`feat/plugin-fronted`分支上为源码插件增加消费端前端托管能力。宿主负责发现、校验和托管插件的`frontend/consumer/`文件，并根据插件生命周期维护挂载索引。

## Manifest Contract

源码插件可以在`plugin.yaml`中声明：

```yaml
consumer:
  frontend:
    mount_path: /portal
    index: index.html
    spa_fallback: true
```

字段语义如下：

| Field | Required | Meaning |
| --- | --- | --- |
| `mount_path` | No | 浏览器可访问的稳定挂载路径，不允许为根路径，不允许覆盖宿主保留前缀。声明即启用。 |
| `index` | No | 入口文件，缺省为`index.html`，路径相对于`frontend/consumer/`。 |
| `spa_fallback` | No | 是否将不存在的干净子路由回退到入口文件；缺省为`false`。 |

保留前缀包括`/api`、`/plugin-assets`、`/consumer-plugin-assets`、`/swagger`、`/api.json`和`/openapi`。

## Asset Hosting

宿主从源码插件的`frontend/consumer/`目录发现消费端前端文件。源码目录和嵌入式文件系统使用同一套发现逻辑，并在清单快照中记录`consumer_frontend`资源计数。

宿主提供两类访问方式：

- 稳定挂载路径：例如`/portal`、`/portal/orders`、`/portal/assets/app.js`，由`consumer.frontend.mount_path`声明。
- 资产命名空间：例如`/consumer-plugin-assets/<plugin-id>/<version>/assets/app.js`，用于按插件 ID 和版本直接读取声明过的消费端前端资源。

稳定挂载路径要求插件已启用。未启用插件、未声明资产或静态资源不存在时返回`404`。`SPA`回退只对插件显式开启`spa_fallback`且请求路径不像静态文件时生效。

## Cache Governance

消费端前端挂载索引是进程内派生缓存，权威来源为源码插件清单、嵌入式或源码目录资产清单、插件 registry 启用状态和版本信息。

一致性模型：

- `cluster.enabled=false`：生命周期操作完成后本进程立即失效挂载索引，下次读取重新构建。
- `cluster.enabled=true`：生命周期操作仍先写入权威状态并标记插件运行时共享修订号；其他实例在读路径观察到运行时修订号变化后，刷新 enabled snapshot、frontend bundle、i18n bundle、WASM 缓存，并清空本地消费端前端挂载索引。

失效触发点：

- 源码插件安装、卸载、启用和禁用。
- 源码插件升级成功。
- 插件运行时共享修订号刷新。

最大可接受陈旧窗口为其他实例下一次读取插件运行时缓存前；读路径会调用运行时缓存新鲜度检查，因此不会依赖本地定时刷新作为唯一一致性机制。失效操作是幂等的，清空本地索引失败不会破坏权威状态，后续读路径可重新构建。

## Governance Snapshot

消费端前端治理快照包含：

- 插件 ID、版本和启用状态。
- 插件租户治理声明，包括是否支持租户治理、scope nature 和默认安装模式。
- 消费端前端挂载路径、入口文件、`SPA`回退状态和资产数量。

## Validation Strategy

- 单元测试覆盖`consumer.frontend`清单校验、资源发现、路径标准化、挂载匹配、`SPA`回退、`HTML base`注入、缓存克隆和治理快照排序。
- HTTP 层测试覆盖插件前端资产响应头、`ETag`协商、消费端资产命名空间解析和既有插件资产路径解析。

## i18n, Permissions, and API Boundaries

本改进不新增用户可见运行时文案，不修改前端语言包、插件`manifest/i18n`资源或`apidoc i18n JSON`。消费端前端资产托管不读取业务表，不改变角色数据权限边界。
