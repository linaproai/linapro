## Why

当前个人中心修改密码场景调用`PUT /api/v1/user/profile`时只提交`password`，但接口 DTO 将`nickname`标记为必填，导致请求在进入服务层前被校验拦截并返回“请输入昵称”。服务层已经按字段执行局部更新，接口校验契约需要与局部更新语义保持一致。

## What Changes

- 调整当前用户资料更新接口校验，使`nickname`、`email`、`phone`、`sex`和`password`都作为可选字段参与局部更新。
- 保持管理员创建用户接口的`nickname`必填约束不变。
- 增加回归测试，覆盖仅提交`password`时`UpdateProfileReq`校验通过。

## Capabilities

### Modified Capabilities

- `user-management`：补充当前用户资料局部更新契约，明确只提交密码时不得因缺少昵称被拒绝。

## Impact

- 后端 API：影响`apps/lina-core/api/user/v1/user_update_profile.go`中的请求 DTO 校验标签。
- 服务层：复用现有`UpdateProfile`按字段更新逻辑，不新增服务方法或运行期依赖。
- 数据权限：当前用户自服务接口仍基于登录上下文更新本人记录，不新增跨用户数据访问路径。
- `i18n`：移除一个已有校验触发点，不新增用户可见文案、菜单、语言包或 API 文档源文本。
- 缓存一致性：不涉及缓存读写或失效策略。
- 开发工具跨平台：不修改脚本、构建工具或跨平台入口。
