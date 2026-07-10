## 1. P0：JSON 载荷政策与 catalog 冻结

- [x] 1.1 将 dedicated codec 治理测试改为方法级冻结名单，并拒绝名单外 dedicated 扩张
- [x] 1.2 在 `pkg/plugin` 中英文 README 写明新增 host service 方法必须使用 JSON envelope

## 2. P0：wire 常量一致性与 JSON helper

- [x] 2.1 在 `hostservices` 单一维护 host service/method wire 常量，catalog 引用常量，治理测试校验一致（不使用 go generate）
- [x] 2.2 保持 `protocol` 对 `HostService*` / `HostServiceMethod*` 的公开 re-export 兼容
- [x] 2.3 删除 `HostServiceCapabilityJSON*` 历史别名，全量切换到 `HostServiceJSON*`
- [x] 2.4 补充/调整 WASM JSON 请求响应共用 helper，降低新增 JSON 方法样板

## 3. P1：upgrade 归 lifecycle 拥有

- [x] 3.1 lifecycle 构造并持有 upgrade 能力，暴露升级预览/执行/源码升级相关方法
- [x] 3.2 根 plugin facade 移除独立 `upgradeSvc` 字段与 `upgrade.New` 装配，改为委托 lifecycle
- [x] 3.3 保持根公开类型别名与错误码兼容，避免管理 API 层破坏性改名

## 4. 验证与文档收口

- [x] 4.1 运行相关包 `go test` 与 `openspec validate simplify-plugin-host-layers --strict`
- [x] 4.2 记录 DI 影响：upgrade 依赖 owner 从根 facade 下沉到 lifecycle，仍复用启动期共享实例

## Feedback

- [x] **FB-1**: 撤销 host service 常量 `go generate` 方案，改为手写 wire 常量 + catalog 一致性测试，降低漏生成风险
- [x] **FB-2**: wire 常量迁入 hostservices 并让 catalog 引用常量，实现单一维护点（无 go generate）
