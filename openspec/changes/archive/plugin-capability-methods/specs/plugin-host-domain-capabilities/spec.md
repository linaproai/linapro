## Requirements

### Requirement: 插件领域能力 Go 方法命名必须避免重复领域名

系统 SHALL 为 `pkg/plugin/capability/*cap` 下的公开 Go 接口使用动作式命名。当 `Service` 所在目录已经表达主资源时，主资源方法 MUST 省略重复领域名，优先使用 `BatchGet`、`Search`、`Current`、`EnsureVisible`、`Delete`、`Run`、`Revoke`、`SetStatus`、`SetEnabled` 等短名。仅当方法目标是子资源、复合对象或短名会造成歧义时，才允许保留限定词。动态 host service 的 wire method 名称 MAY 保持显式原样，不受本规则影响。
