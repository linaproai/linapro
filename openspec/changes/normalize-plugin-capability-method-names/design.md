## Context

插件能力目录里的 `Service` 接口已经把主资源表达得很明确，但现有方法名仍常重复领域名，例如 `BatchGetUsers`、`SearchSessions`、`EnsurePluginsVisible`。这让插件开发者在调用面和实现面都要反复写同一个名词，也会让新增方法看起来比实际复杂。

这次变更只调整 typed Go 能力接口与实现命名，不改动态 `host service` wire method 名称，不改数据权限语义，不改缓存语义，也不引入兼容层。项目没有历史兼容负担，所以可以一次性重命名并让编译门禁覆盖所有遗漏点。

## Goals / Non-Goals

**Goals:**

- 让主资源能力方法更短、更统一，减少插件调用噪音。
- 让“是否是主资源方法”一眼可见，避免 `Users.SearchUsers` 这种重复表达。
- 保持子资源或歧义场景的限定词，避免过度压缩导致语义模糊。
- 只改 typed Go 接口与其实现，不改动态协议字符串。

**Non-Goals:**

- 不新增业务能力。
- 不改 `users.batch_get`、`messages.batch_get` 这类 wire method 名称。
- 不为旧方法保留兼容别名。
- 不重命名内部 `datascope`、`tenantspi` 这类宿主私有辅助接口。

## Decisions

1. **主资源能力方法改为动作名。**
   当 `usercap`、`filecap`、`jobcap`、`sessioncap`、`plugincap`、`notifycap` 这类目录已经表达主资源时，公开方法使用动作名即可，例如 `BatchGet`、`Search`、`EnsureVisible`、`Delete`、`Run`、`Revoke`、`SetStatus`、`SetEnabled`。
   这样比 `BatchGetUsers`、`SearchUsers` 更短，也更符合“目录表达名词，方法表达动作”的分层。

2. **子资源和歧义方法保留限定词。**
   对象不是主资源本身、或者短名会丢失语义时，继续保留限定词，例如 `BatchGetPermissions`、`ListUserTenants`、`SearchDepartments`、`DeleteBySource`。
   这样可以避免把通用规则用成过度压缩。

3. **动态协议字符串不跟着改名。**
   `pluginbridge` 的 wire method 名称继续保持显式资源名，例如 `users.batch_get`。
   原因是 wire method 主要承担 transport 和授权治理职责，显式命名更利于审计和协议稳定。

4. **不加兼容层，直接一次性改名。**
   这次变更面向全新项目，没有历史兼容包袱。保留旧方法只会让新旧命名长期并存，降低规则清晰度。
   由编译和测试门禁直接暴露遗漏调用点更直接。

## Risks / Trade-offs

- [Risk] 触及文件多，漏改任一调用点会导致编译失败。→ [Mitigation] 用机械重命名配合全量 `go test` 和静态检索收口。
- [Risk] 文档与代码命名不一致。→ [Mitigation] 同步更新 `localdocs`、`README`、OpenSpec 和后端规则。
- [Risk] 过度压缩后部分方法语义不够明确。→ [Mitigation] 对子资源、复合目标和运维命令保留限定词。

## Migration Plan

1. 先改 `pkg/plugin/capability` 公开接口和对应 host adapter。
2. 再改 `pluginbridge`、`WASM` 分发器和插件调用点。
3. 同步更新测试替身、README、设计文档和项目规则。
4. 用编译门禁和 `openspec validate` 收口，确认没有旧方法残留。

回滚策略很简单：如果验证失败，直接回滚本次重命名改动即可，因为没有兼容层和数据迁移。

## Open Questions

- 无。主资源方法统一用动作名，子资源方法保留限定词，动态 wire method 保持显式原名。
