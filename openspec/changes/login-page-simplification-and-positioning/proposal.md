## Why

当前登录页仍然暴露了忘记密码、注册账号、手机号登录、扫码登录和其他登录方式等未落地入口，容易给用户造成“系统已支持但不可用”的误导。同时，登录框默认位置需要统一为居右展示，登录说明文案也需要更新为更贴合产品定位的表达，并允许后续通过系统参数进行运营级调整，因此需要为登录页展示策略补齐一轮正式迭代。

## What Changes

- 收敛登录页默认可见能力，仅保留当前已落地的账号密码登录入口，隐藏忘记密码、注册账号、手机号登录、扫码登录和其他第三方登录方式。
- 将登录框默认位置调整为居右展示，并梳理前端现有可调节的登录页布局能力，统一落到 LinaPro 的登录页配置链路中。
- 更新登录页默认说明文案，使其更贴合业务演进与插件扩展能力定位。
- 在宿主 `sys_config` 的 public frontend 配置体系中新增登录框位置参数，使管理员可以通过系统参数维护登录页默认布局。
- 扩展公共前端配置白名单响应与前端运行时设置同步逻辑，让未登录页面也能读取并应用登录框位置配置。
- 为登录页展示行为与系统参数治理补充对应规格和实现任务，确保后续增加更多登录方式时有明确的开关策略。

## Capabilities

### New Capabilities
- `login-page-presentation`: 定义登录页在当前阶段仅暴露账号密码登录入口、隐藏未实现入口，以及登录框默认居右且支持位置配置的展示行为。

### Modified Capabilities
- `config-management`: 扩展内置 public frontend 系统参数元数据与校验规则，新增登录框位置配置并通过公共前端配置接口暴露给登录页消费。

## Impact

- **前端页面**: 影响 `apps/lina-vben/apps/web-antd/src/views/_core/authentication/`、`apps/lina-vben/apps/web-antd/src/layouts/auth.vue` 与公共前端运行时配置同步逻辑。
- **前端共享组件**: 可能影响 `apps/lina-vben/packages/effects/common-ui/src/ui/authentication/` 下的登录页组件入参或默认行为。
- **后端配置服务**: 影响 `apps/lina-core/internal/service/config/` 的 public frontend 配置模型、默认值与参数校验。
- **系统参数管理**: 影响宿主 `sys_config` 内置参数清单，以及配置管理页对新增受保护参数的展示与维护。
