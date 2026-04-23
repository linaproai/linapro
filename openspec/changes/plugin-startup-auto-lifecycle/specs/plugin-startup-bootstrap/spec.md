## ADDED Requirements

### Requirement: 宿主必须在主配置文件中提供简化的插件自动启用配置
系统 SHALL 在宿主服务主配置文件中提供 `plugin.autoEnable` 配置项，并使用插件 ID 列表声明哪些插件需要在系统启动时自动启用。该配置 MUST 不要求用户继续填写 `desiredState`、`required` 或授权明细等复杂结构。

#### Scenario: 解析有效自动启用列表
- **WHEN** 宿主启动并读取主配置文件中的 `plugin.autoEnable` 数组
- **THEN** 宿主成功构建需要自动启用的插件 ID 集合
- **AND** 该集合中的每个元素都按插件 ID 语义解析

#### Scenario: 拒绝非法自动启用配置
- **WHEN** `plugin.autoEnable` 不是字符串数组，或数组中存在空字符串插件 ID
- **THEN** 宿主 MUST 拒绝继续启动
- **AND** 错误信息 MUST 明确指出 `plugin.autoEnable` 配置非法

### Requirement: 宿主必须在插件接线之前执行启动期 bootstrap
系统 SHALL 在插件 HTTP 路由注册、插件 cron 接线和动态前端 bundle 预热之前，先按 `plugin.autoEnable` 推进命中插件的生命周期状态。

#### Scenario: 启动前推进源码插件到启用态
- **WHEN** 已发现的源码插件出现在 `plugin.autoEnable` 列表中
- **THEN** 宿主在插件路由与插件 cron 注册前先完成该源码插件的安装与启用
- **AND** 随后的 enabled snapshot 读取结果中该插件处于启用状态

#### Scenario: 未列入自动启用列表的插件保持人工治理
- **WHEN** 插件被宿主发现，但没有出现在 `plugin.autoEnable` 列表中
- **THEN** 宿主仅执行常规 manifest 同步与注册表刷新
- **AND** 宿主 MUST NOT 因启动 bootstrap 而自动安装或自动启用该插件

### Requirement: 自动启用列表必须隐式包含安装与启用语义
系统 SHALL 将 `plugin.autoEnable` 中的插件解释为“宿主启动时必须处于启用状态”的插件；若插件尚未安装，宿主 MUST 先完成安装，再继续启用。

#### Scenario: 自动启用首次发现的源码插件
- **WHEN** 源码插件出现在 `plugin.autoEnable` 列表中，且当前仍处于未安装、未启用状态
- **THEN** 宿主先完成该源码插件安装
- **AND** 随后继续把它推进到启用状态

#### Scenario: 已启用插件重复命中自动启用列表
- **WHEN** 某插件当前已经处于启用状态，且其插件 ID 仍然存在于 `plugin.autoEnable` 列表中
- **THEN** 宿主保持该插件启用状态不变
- **AND** 宿主 MUST NOT 因自动启用流程重复执行而把它回退到更低状态

### Requirement: 自动启用列表命中的插件失败时必须阻止宿主启动
系统 SHALL 将 `plugin.autoEnable` 视为宿主启动期的显式必需插件列表；列表中的任一插件缺失、安装失败、启用失败或未在等待窗口内达到启用态时，宿主 MUST fail-fast。

#### Scenario: 自动启用插件缺失导致启动失败
- **WHEN** `plugin.autoEnable` 中声明了某插件 ID，但宿主在启动期无法发现该插件
- **THEN** 宿主 MUST 终止启动流程
- **AND** 返回的错误信息 MUST 包含缺失插件的插件 ID

#### Scenario: 自动启用插件启用失败导致启动失败
- **WHEN** `plugin.autoEnable` 中声明的某插件在安装、启用或等待收敛阶段失败
- **THEN** 宿主 MUST 终止启动流程
- **AND** 错误信息 MUST 包含失败插件标识与失败阶段

### Requirement: 集群模式下启动期 bootstrap 必须区分共享生命周期动作与本地收敛
系统 SHALL 在集群模式下仅允许主节点执行插件共享生命周期动作（如安装 SQL、菜单写入、release 切换与共享状态推进），从节点只等待共享状态结果并刷新本地投影。

#### Scenario: 主节点执行共享插件动作
- **WHEN** 集群模式下某插件出现在 `plugin.autoEnable` 列表中，且需要执行安装或启用推进
- **THEN** 只有主节点执行该插件的共享安装、启用或 reconcile 动作
- **AND** 从节点 MUST NOT 重复执行同一插件的共享副作用步骤

#### Scenario: 从节点等待共享状态收敛后刷新本地视图
- **WHEN** 集群模式下从节点启动并发现某插件出现在 `plugin.autoEnable` 列表中
- **THEN** 从节点等待主节点写入共享稳定状态或等待窗口超时
- **AND** 随后基于共享结果刷新本地 enabled snapshot 与运行时投影

### Requirement: 启动期自动启用动态插件时必须复用既有授权快照
系统 SHALL 在动态插件出现在 `plugin.autoEnable` 列表中且声明了受治理 host services 时，复用当前 release 已确认的授权快照；宿主 MUST NOT 要求用户在主配置文件中填写复杂授权明细。

#### Scenario: 复用既有授权快照自动启用动态插件
- **WHEN** 动态插件出现在 `plugin.autoEnable` 列表中，且其当前 release 已存在宿主确认过的授权快照
- **THEN** 宿主复用该授权快照推进动态插件自动启用
- **AND** 宿主 MUST NOT 再要求从主配置文件读取授权明细

#### Scenario: 缺少授权快照时拒绝自动启用动态插件
- **WHEN** 动态插件出现在 `plugin.autoEnable` 列表中、声明了受治理 host services，但当前 release 尚未形成授权快照
- **THEN** 宿主 MUST 终止启动
- **AND** 错误信息 MUST 明确指出需要先通过常规审核流程生成授权快照

### Requirement: 插件管理界面必须标识启动自动启用插件并提示临时治理后果
系统 SHALL 在插件管理列表与详情视图中，以只读方式标识当前插件是否被宿主主配置文件中的 `plugin.autoEnable` 命中；当管理员在界面中对这类插件执行禁用或卸载时，界面 MUST 在请求提交前明确说明“本次操作立即生效，但若配置不变，宿主下次重启后会再次安装并启用该插件”。

#### Scenario: 列表与详情展示启动自动启用标识
- **WHEN** 某插件当前插件 ID 存在于宿主主配置文件的 `plugin.autoEnable` 列表中
- **THEN** 插件管理列表 SHALL 展示该插件由 `plugin.autoEnable` 管理的只读标识
- **AND** 插件详情视图 SHALL 展示相同语义的只读说明

#### Scenario: 禁用启动自动启用插件前提示重启后果
- **WHEN** 管理员尝试在插件管理页面禁用一个命中 `plugin.autoEnable` 的插件
- **THEN** 界面 MUST 在真正发起禁用请求前先展示风险确认提示
- **AND** 提示内容 MUST 明确指出“本次禁用立即生效，但若 `plugin.autoEnable` 配置不变，则宿主重启后会再次启用该插件”
- **AND** 提示内容 MUST 明确指出“若要永久停用，需要先修改宿主主配置文件中的 `plugin.autoEnable`”

#### Scenario: 卸载启动自动启用插件时提示重启后果
- **WHEN** 管理员尝试在插件管理页面卸载一个命中 `plugin.autoEnable` 的插件
- **THEN** 卸载确认界面 MUST 展示风险提示
- **AND** 提示内容 MUST 明确指出“本次卸载立即生效，但若 `plugin.autoEnable` 配置不变，则宿主重启后会再次安装并启用该插件”
- **AND** 提示内容 MUST 明确指出“若要永久停用，需要先修改宿主主配置文件中的 `plugin.autoEnable`”
