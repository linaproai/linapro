## MODIFIED Requirements

### Requirement:翻译查找热路径必须避免克隆整个运行时消息包

宿主系统 SHALL 让 `Translate` 等单值返回的翻译查找方法在缓存命中时直接持有内部消息包的读锁并查找值，不得克隆或复制整个运行时消息包。仅当方法语义上需要向调用方返回消息集时（如运行时翻译包 API），系统可在返回前克隆一次。源码文案兜底、key 兜底和默认语言上下文兜底应通过 `Translate(ctx, key, fallback)` 和调用方传入的上下文或 fallback 表达，不应为每种兜底语义新增独立 service 方法。

#### Scenario:单键翻译查找在缓存命中时不克隆整个消息包

- **当** 业务模块调用 `Translate(ctx, key, fallback)` 且当前语言运行时消息缓存已存在
- **则** 系统仅持有内部消息包的读锁并查找值，直接返回找到的字符串
- **且** 不执行 `cloneFlatMessageMap` 或等效的完整 `map[string]string` 复制
- **且** 调用方仍收到与之前语义一致的结果

#### Scenario:翻译包在外部返回时保留克隆语义

- **当** 控制器调用 `BuildRuntimeMessages` 向前端返回消息集时
- **则** 系统在移交消息集前克隆一次，确保调用方可安全独立持有
- **且** 该克隆不损坏或覆盖内部缓存
