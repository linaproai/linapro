## Context

`Storage()`已经被明确为插件私有对象存储领域能力，源码插件通过同进程`storagecap.Service.Put(ctx, PutInput{Body: io.Reader})`直接流式写入宿主 provider。动态插件目前经过`pluginbridge`transport，`domainhostcall`会先`io.ReadAll`，再用单次`storage.put`把完整`[]byte`传给 WASM host service，这会把最终对象大小转化为 guest 内存和单次 host call payload 压力。

本变更属于动态插件 transport 能力增强，不改变`storagecap.Service`作为领域 owner 的职责，也不让`Files()`承担插件私有对象写入。

## Goals / Non-Goals

**Goals:**

- 动态插件`Storage().Put`支持小文件单次上传和大文件/未知大小分片上传。
- 分片上传过程中宿主内存占用受单个 chunk 限制，不随最终对象大小线性增长。
- 分片提交时继续进入`storagecap.Service.Put`，复用插件 ID、租户、provider、content type、overwrite 和对象 key 隔离语义。
- 每个分片相关 host call 都校验同一个最终 logical path 的`storage.resources.paths`授权，并校验 upload ID 与路径绑定关系。
- 实现范围保持在动态插件 protocol、guest adapter 和 WASM storage host service 内。

**Non-Goals:**

- 不新增 HTTP 上传 API、预签名 URL 或 provider 直传能力。
- 不为源码插件新增分片协议；源码插件已经能通过 Go `io.Reader`直接流式写入。
- 不改变宿主文件中心`Files()`领域、不创建`sys_file`记录、不引入文件中心下载语义。
- 不新增数据库表或持久化跨节点上传会话。

## Decisions

1. 保留`storage.put`并新增分片方法，而不是把所有上传都改成分片。

   小文件单次上传路径简单、已有测试覆盖，也便于保持当前 typed SDK 行为；大文件和未知大小 reader 才进入`put.init`、`put.chunk`、`put.commit`、`put.abort`。这样调用方仍只使用`Storage().Put`领域方法，传输策略由 SDK 内部选择。

2. 上传会话由 WASM storage host service 管理，最终提交回到`storagecap.Service.Put`。

   会话只保存 upload ID、插件 ID、logical path、content type、overwrite、临时文件路径、当前 offset 和过期时间。宿主不在分片阶段写最终 provider 对象，只有 commit 时以`os.File`作为`io.Reader`调用领域服务，避免半成品对象进入最终存储空间。

3. 分片 payload 有上限，最终对象没有固定 host-service 上限。

   分片上限用于保护单次 WASM host call 的内存边界；最终对象大小不由 bridge 硬编码限制，继续交给 provider、存储空间策略或业务层处理。这与前一轮去除`storagecap.Service`固定对象大小限制的方向一致。

4. 上传会话保持进程内和临时文件语义。

   动态插件一次`Storage().Put`调用是同步 SDK 流程，init、chunk、commit 在同一宿主 runtime 实例内完成；进程内会话足够支撑该流程。集群或断点续传需要持久化会话和跨节点协调，本变更不引入这类产品语义。

5. 分片授权绑定最终 logical path。

   `put.init`、`put.chunk`、`put.commit`和`put.abort`请求都携带最终 logical path，host service 分发器继续使用该 path 做`storage.resources.paths`授权。宿主会话额外校验 upload ID、plugin ID 和 path 一致，避免插件把某个已授权会话挪用到其他对象。

## Risks / Trade-offs

- [Risk] 上传进程崩溃会留下临时文件。→ 会话使用独立临时目录和 TTL 清理；commit 和 abort 都删除临时文件。
- [Risk] 多实例部署下分片请求如果跨节点会找不到 upload ID。→ guest SDK 发起同步 host call 序列，当前 WASM runtime 调用链不会主动跨节点；设计中明确不提供跨节点断点续传语义。
- [Risk] 分片 offset 错误可能导致对象内容错乱。→ 宿主只接受等于当前 offset 的顺序分片，commit 校验总大小与累计 offset 一致。
- [Risk] 分片方法扩大动态插件 host service 攻击面。→ 每个方法都执行 service、method、resource path 授权，并对 upload ID、路径、插件 ID、chunk 大小和会话过期做校验。

## Migration Plan

动态插件业务代码不需要迁移，继续调用`Storage().Put`。已有小文件调用继续走`storage.put`。大文件或未知大小 reader 会在新 SDK 中自动走分片上传；需要使用该路径的动态插件必须在`plugin.yaml hostServices`中声明`put.init`、`put.chunk`、`put.commit`和`put.abort`方法。示例动态插件同步补充这些声明。

## Open Questions

无。
