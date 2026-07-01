## ADDED Requirements

### Requirement: 当前用户资料更新必须支持局部字段更新

系统 SHALL 允许当前登录用户通过`PUT /api/v1/user/profile`按字段局部更新个人资料。`nickname`、`email`、`phone`、`sex`和`password`均为可选更新字段；调用方只提交其中一个字段时，接口校验 MUST NOT 因其他字段缺失而拒绝请求。服务层 MUST 仅更新请求中显式提交的字段，并保持未提交字段原值不变。

#### Scenario: 仅提交密码时更新个人资料

- **WHEN** 当前登录用户调用`PUT /api/v1/user/profile`
- **AND** 请求体只包含`password`
- **THEN** 请求校验通过
- **AND** 系统更新当前用户密码
- **AND** 不要求调用方同时提交`nickname`

#### Scenario: 仅提交昵称时更新个人资料

- **WHEN** 当前登录用户调用`PUT /api/v1/user/profile`
- **AND** 请求体只包含`nickname`
- **THEN** 请求校验通过
- **AND** 系统更新当前用户昵称
- **AND** 未提交的邮箱、手机号、性别和密码保持原值
