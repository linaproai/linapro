# 演示控制守卫

## 目的

定义官方 `demo-control` 源码插件在启用演示模式时如何启用只读演示保护，同时保留必要的会话访问并阻断插件治理写操作。

## 需求

### 需求：演示只读模式由插件的启用状态控制系统

系统必须将 `demo-control` 的已安装并启用状态视为演示保护的运行时开关。`plugin.autoEnable` 仅控制启动安装和启用；启动后不得将其视为单独的运行时开关。

#### 场景：默认配置下演示保护保持禁用
- **当** 宿主以默认交付配置启动且 `plugin.autoEnable` 不包含 `demo-control` 时
- **则** 宿主不安装或启用 `demo-control`
- **且** 从未启用该插件的部署默认不阻断写操作

#### 场景：手动启用激活演示保护
- **当** 管理员安装并启用 `demo-control` 时
- **则** demo-control 中间件对后续请求生效
- **且** 写请求被只读演示规则阻断

### 需求：宿主必须随源码树交付 demo-control 源码插件

系统必须交付名为 `demo-control` 的官方源码插件，使部署可通过启动配置或插件治理启用该能力。

#### 场景：宿主发现 demo-control 源码插件
- **当** 宿主扫描源码插件并同步注册表数据时
- **则** 发现 `demo-control`
- **且** 运维人员可决定是否启用

### 需求：demo-control 插件启用时必须阻断系统写操作

启用时，demo-control 必须按 HTTP 方法语义阻断系统写请求，同时允许读式请求。

#### 场景：禁用时无写拦截
- **当** `demo-control` 未启用时
- **则** `POST`、`PUT` 和 `DELETE` 请求不被 demo-control 拒绝

#### 场景：查询式请求保持允许
- **当** `demo-control` 已启用
- **且** 请求使用 `GET`、`HEAD` 或 `OPTIONS` 时
- **则** demo-control 允许请求继续

#### 场景：写请求被拒绝
- **当** `demo-control` 已启用
- **且** 请求使用 `POST`、`PUT` 或 `DELETE` 时
- **则** demo-control 以清晰的只读演示消息拒绝请求
- **且** 请求不继续进入业务处理

### 需求：demo-control 插件必须保留最小会话白名单

系统必须在 demo-control 启用时保留登录和退出行为，使演示环境保持可用。

#### 场景：登录保持允许
- **当** `demo-control` 已启用
- **且** 请求为 `POST /api/v1/auth/login` 时
- **则** demo-control 允许请求继续

#### 场景：退出保持允许
- **当** `demo-control` 已启用
- **且** 请求为 `POST /api/v1/auth/logout` 时
- **则** demo-control 允许请求继续

### 需求：demo-control 插件启用时必须拒绝插件治理写操作

`demo-control` 启用时，系统 SHALL 拒绝插件治理写操作，包括插件同步、动态包上传、安装、卸载、启用和禁用。插件管理的 `GET`、`HEAD` 和 `OPTIONS` 请求作为只读操作保持允许。

#### 场景：启用 demo-control 时拒绝插件安装
- **当** `demo-control` 已启用
- **且** 请求为 `POST /api/v1/plugins/{id}/install` 时
- **则** demo-control 以只读演示消息拒绝请求

#### 场景：拒绝插件启用和禁用请求
- **当** `demo-control` 已启用
- **且** 请求为 `PUT /api/v1/plugins/{id}/enable` 或 `PUT /api/v1/plugins/{id}/disable` 时
- **则** demo-control 以只读演示消息拒绝请求

#### 场景：拒绝插件卸载
- **当** `demo-control` 已启用
- **且** 请求为 `DELETE /api/v1/plugins/{id}` 时
- **则** demo-control 以只读演示消息拒绝请求

#### 场景：拒绝插件同步和上传写操作
- **当** `demo-control` 已启用
- **且** 请求为 `POST /api/v1/plugins/sync` 或 `POST /api/v1/plugins/dynamic/package` 时
- **则** demo-control 以只读演示消息拒绝请求

#### 场景：插件管理读取保持允许
- **当** `demo-control` 已启用
- **且** 请求为使用 `GET`、`HEAD` 或 `OPTIONS` 的插件管理查询时
- **则** demo-control 允许请求继续
