# multi-tenant

`multi-tenant` 是 LinaPro 官方提供的多租户源码插件。

[English](README.md) | 简体中文

## 能力范围

该插件负责：

- 租户增删改查与生命周期状态迁移
- 用户成员关系管理
- 代码内置的租户解析策略
- 租户生命周期前置保护与显式插件治理初始化
- 平台管理员 impersonation 与 LifecycleGuard 前置校验

## 宿主边界

宿主只保留`tenantcap`、租户中间件、`bizctx`租户身份、租户感知缓存作用域和插件治理字段等稳定接缝。租户主体、成员关系、解析策略、生命周期保护和租户管理工作流都收敛在本插件内。

未安装或未启用本插件时，宿主回退到`tenant_id = 0`，保持单租户开箱体验。

## 配置

多租户默认策略由代码维护，不再通过宿主配置模板维护：

| 默认项 | 默认值 | 含义 |
|--------|--------|------|
| 隔离模型 | `pool` | 基于`tenant_id`列的`Pool`模型。 |
| 用户租户基数 | `multi` | 允许一个用户加入多个租户。 |
| 解析器链 | `override,jwt,session,header,subdomain,default` | 租户解析器顺序。 |
| 歧义处理 | `prompt` | 无法明确判断租户时返回租户选择要求。 |
| 根域名 | 空 | 预留给后续设置入口，当前禁用子域名解析。 |

首版解析策略固定在代码中，不提供解析配置表、运行时修改路径或分布式解析策略广播。`rootDomain`当前保持为空，待后续设置设计明确开放后再引入。

## 插件治理

本插件使用`platform_only`、`supports_multi_tenant: false`和`global`，因为它自身承载租户控制面。租户感知业务插件应使用`scope_nature: tenant_aware`，在支持租户级治理时声明`supports_multi_tenant: true`，并按需要选择`global`或`tenant_scoped`安装模式。
