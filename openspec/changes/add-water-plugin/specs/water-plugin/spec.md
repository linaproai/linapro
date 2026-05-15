## ADDED Requirements

### Requirement: 水印源码插件注册
系统 SHALL 提供 `water` 源码插件，并通过插件清单注册菜单、权限、前端页面和后端路由。

#### Scenario: 插件随宿主注册
- **WHEN** 宿主启动并加载源码插件
- **THEN** `water` 插件的 `/water` 页面和 `/api/v1/water/*` 接口可通过插件治理链路访问

#### Scenario: 插件不使用运行时 i18n
- **WHEN** 用户打开 water 插件页面
- **THEN** 页面文案直接展示中文，不依赖运行时 i18n 翻译键

### Requirement: 水印截图任务提交
系统 SHALL 提供截图水印任务提交接口，接收设备、租户、图片和可选回调地址，并立即返回任务 ID。

#### Scenario: 成功提交任务
- **WHEN** 调用方提交包含媒体租户 ID 和 base64 图片的截图请求
- **THEN** 系统创建水印任务并返回 `success=true` 与稳定 `taskId`

#### Scenario: 参数缺失
- **WHEN** 调用方缺少媒体租户 ID 或图片
- **THEN** 系统拒绝请求并返回稳定业务错误码

### Requirement: Media 策略集成
系统 SHALL 从 `media_*` 表解析水印策略，旧 hotgo `hg_*` 表名在 LinaPro 中 MUST 映射为对应 `media_*` 表。

#### Scenario: 按优先级解析策略
- **WHEN** 同时存在租户设备、设备、租户和全局策略
- **THEN** 系统按租户设备、设备、租户、全局的顺序选择第一个启用策略

#### Scenario: 未安装 media 表
- **WHEN** media 策略表不存在或尚未安装
- **THEN** 系统返回稳定业务错误，不创建 `hg_*` 表

### Requirement: 水印配置解析
系统 SHALL 解析策略 YAML 中的 `watermark` 配置，并兼容 hotgo 旧字段语义。

#### Scenario: 解析水印节点
- **WHEN** media 策略 YAML 包含 `watermark.enabled=true` 和文字配置
- **THEN** 系统使用该配置绘制文字水印

#### Scenario: 未启用水印
- **WHEN** 匹配策略未包含水印配置或 `enabled=false`
- **THEN** 系统跳过绘制并把任务标记为跳过

### Requirement: 图片水印处理
系统 SHALL 支持 PNG/JPEG base64 输入，并输出 PNG data URL 格式的水印图片。

#### Scenario: 同步预览
- **WHEN** 管理页提交同步预览请求
- **THEN** 系统返回处理后的图片、策略来源和耗时

#### Scenario: 无效图片
- **WHEN** 调用方提交非法 base64 或无法解码的图片
- **THEN** 系统返回稳定业务错误或将异步任务标记为失败

### Requirement: 任务状态查询
系统 SHALL 保存最近水印任务状态，并允许按任务 ID 查询。

#### Scenario: 查询任务结果
- **WHEN** 调用方使用已返回的 `taskId` 查询任务
- **THEN** 系统返回任务状态、策略来源、错误信息和输出图片信息

#### Scenario: 查询不存在任务
- **WHEN** 调用方查询不存在的任务 ID
- **THEN** 系统返回稳定业务错误码

### Requirement: 回调通知
系统 SHALL 在异步任务完成后向请求中的回调地址发送处理结果。

#### Scenario: 回调成功
- **WHEN** 请求包含有效回调地址且水印处理完成
- **THEN** 系统以 POST JSON 形式发送处理结果，并记录任务为成功或跳过

#### Scenario: 回调失败
- **WHEN** 回调地址不可达或返回非 2xx
- **THEN** 系统记录回调错误并保留任务查询信息
