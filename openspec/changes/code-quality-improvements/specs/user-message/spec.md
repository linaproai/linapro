## ADDED Requirements

### Requirement: 用户消息轮询必须支持页面可见性感知
用户消息未读数轮询 SHALL 监听页面可见性事件。当 `document.visibilityState === 'hidden'` 时 MUST 暂停轮询；当页面再次可见时 MUST 立即触发一次未读数刷新并恢复周期性轮询。底层定时器 MUST 在用户登出或 store 销毁时显式停止。

#### Scenario: 页面隐藏时暂停未读数轮询
- **WHEN** 用户切换到其他标签页
- **THEN** 消息 store MUST 暂停未读数轮询
- **AND** 不再产生 `GET /api/v1/user/message/count` 请求

#### Scenario: 页面恢复可见时立即刷新一次
- **WHEN** 用户从其他标签页切回当前应用
- **THEN** 消息 store MUST 立即触发一次未读数刷新
- **AND** 之后恢复周期性轮询

#### Scenario: 用户登出时停止轮询定时器
- **WHEN** 用户执行登出操作
- **THEN** 消息 store MUST 显式停止轮询定时器
- **AND** 不再产生未读数请求
