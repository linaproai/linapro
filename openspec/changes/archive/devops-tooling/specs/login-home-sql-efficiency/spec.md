## ADDED Requirements

### Requirement: 登录后首页运行期 SQL 必须避免可复用元数据重复读取

登录后首页加载涉及用户信息、菜单、插件状态、消息和租户等多个接口。系统 SHALL 避免在同一请求、同一列表投影或同一可复用读取作用域内重复读取插件 catalog 等小型运行期元数据。

#### Scenario: 插件 release 行在同一读取作用域内复用

- **WHEN** 一个首页相关请求或插件列表投影需要多次解析同一插件 release
- **THEN** 系统必须复用已读取的 release 行或批量读取结果
- **AND** 不得对同一 `plugin_id + release_version` 或同一 release ID 反复执行等价数据库查询

### Requirement: 鉴权会话校验必须减少每请求数据库往返

系统 SHALL 在保持强制下线、租户隔离和会话超时语义不变的前提下，减少每个鉴权请求对在线会话表的重复 SQL 往返。

#### Scenario: 有效会话使用单次读取判断状态

- **WHEN** 请求携带有效 token、租户匹配且会话未过期
- **THEN** 系统应通过单次会话记录读取判断有效性
- **AND** 仅当 `last_active_time` 超出写入节流窗口时才更新在线会话记录

### Requirement: 登录后首页 SQL 优化必须有自动化验证

登录后首页 SQL 优化属于运行期行为变更。系统 SHALL 通过自动化测试验证优化后的关键行为。
