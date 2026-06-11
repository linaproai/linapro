## 1. 规范和协议设计

- [x] 1.1 更新动态插件`AI`host service 增量规范，移除`resources`授权模型并定义方法授权模型
- [x] 1.2 更新设计记录，明确宿主边界、`i18n`、缓存、数据权限、测试和 DI 影响

## 2. 核心实现

- [x] 2.1 将`ai`host service 目录调整为不接受`resources`，并拒绝`ai.resources`清单声明
- [x] 2.2 更新 guest SDK，使`AI`调用 envelope 不再携带`purpose`资源引用
- [x] 2.3 更新`WASM` host handler，移除`purpose`资源匹配和`resources.attributes`策略校验，仅保留 service、method、DTO 和来源身份治理
- [x] 2.4 更新`apps/lina-core/pkg/plugin` README 中的`ai`声明示例，并静态确认动态插件示例清单不含旧`ai.resources`声明

## 3. 验证和审查

- [x] 3.1 更新并运行`pluginbridge`和`WASM` host service 单元测试，覆盖`ai.resources`拒绝、方法授权成功和未授权方法拒绝
- [x] 3.2 运行 Go 编译门禁和`openspec validate simplify-dynamic-ai-host-service-auth --strict`
- [x] 3.3 执行`lina-review`，确认 OpenSpec、插件、后端 Go、测试、文档和`i18n`影响均已闭环

## Feedback

- [x] **FB-1**: 删除动态插件 README 中`secret.resolve`、`event.publish`和`queue.enqueue`预留治理条目说明
- [x] **FB-2**: 补齐动态插件`service: user`用户领域读取和可见性校验 host service
- [x] **FB-3**: 补齐动态插件普通领域 host service，并将重叠动态能力收敛到`capability.Services`领域能力调用路径
- [x] **FB-4**: 复用`capability/*cap`领域接口，消除动态插件`guest`领域接口重复定义
- [x] **FB-5**: 将动态插件`guest`领域 hostcall 实现迁移到内部子组件，降低公开`guest`包文件膨胀
- [x] **FB-6**: 将`guest/internal/domainhostcall`普通领域 hostcall 实现继续按领域能力拆分维护
- [x] **FB-7**: 移除动态插件`guest`领域服务别名，直接引用`capability/*cap`领域接口类型
- [x] **FB-8**: 将`WASM`普通领域 host service 分发从单一大文件拆分为按领域职责维护的文件
