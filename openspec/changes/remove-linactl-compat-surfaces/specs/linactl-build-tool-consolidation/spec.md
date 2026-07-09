## MODIFIED Requirements

### Requirement: 插件自定义构建指令必须从插件根配置读取

系统 SHALL 允许源码插件和动态插件在插件根`hack/config.yaml`中通过`build.commands`声明自定义构建指令。仓库根目录执行`make build`或`linactl build`且未指定`dir=`时，系统 MUST 遍历`apps/lina-plugins`下所有包含`plugin.yaml`的直接插件目录并执行对应插件的自定义构建指令；指定`dir=apps/lina-plugins/<plugin-id>`时，系统 MUST 只执行该插件的构建。插件自定义构建指令 MUST 在插件根目录中执行，并且变量展开 MUST 只包含`$(TARGET_DIR)`、`$(BUILD_DIR)`和`$(REPO_ROOT)`。非插件或插件定向构建目标 MUST 以`hack/config.yaml`作为唯一自定义命令来源；目标目录缺少`hack/config.yaml`时不得回退读取`package.json`构建脚本。

#### Scenario: 根目录构建遍历所有插件

- **WHEN** 开发者在仓库根目录运行`make build`且未指定`dir=`
- **THEN** 系统构建宿主前端、宿主打包资源和宿主后端
- **AND** 遍历`apps/lina-plugins`下所有直接插件目录
- **AND** 对每个包含`plugin.yaml`的插件执行其`hack/config.yaml`中的`build.commands`

#### Scenario: 定向构建只执行指定插件

- **WHEN** 开发者运行`make build dir=apps/lina-plugins/<plugin-id>`
- **THEN** 系统只执行该插件`hack/config.yaml`中的`build.commands`
- **AND** 不执行其他插件的构建指令

#### Scenario: 插件未声明自定义构建指令

- **WHEN** 插件根`hack/config.yaml`不存在或未声明`build.commands`
- **THEN** 系统跳过该插件的自定义构建指令
- **AND** 不将缺少自定义构建指令视为错误

#### Scenario: 仅标准构建变量被展开

- **WHEN** 插件或定向构建目标的`hack/config.yaml`中使用`$(TARGET_DIR)`、`$(BUILD_DIR)`或`$(REPO_ROOT)`
- **THEN** 系统按标准变量集合展开这些值
- **AND** 不包含旧变量名的专门识别或处理分支

#### Scenario: 定向构建目标没有配置文件

- **WHEN** 开发者运行`linactl build dir=<target>`
- **AND** `<target>/hack/config.yaml`不存在
- **THEN** 系统拒绝该定向构建
- **AND** 不读取`<target>/package.json`中的`build`脚本

### Requirement: `linactl`公开参数必须只接受标准契约

系统 SHALL 只接受当前标准化的`linactl`参数和配置输入。显式插件模式参数 MUST 使用标准布尔值，文档示例 SHOULD 使用`plugins=0`或`plugins=1`；未传入`plugins`时，系统 MAY 根据官方插件工作区是否存在插件清单自动选择宿主模式或插件完整模式。系统 MUST reject 显式`plugins=auto`，且不得为该旧值保留专门判断分支。公开命令参数`key`MUST 使用文档中声明的`kebab-case`形式，系统不得将`snake_case key`自动映射为`kebab-case key`，也不得将`kebab-case key`自动映射为`snake_case key`；根`Makefile`包装入口也 MUST 只转发当前标准参数名。动态`WASM`单插件构建的显式源码目录 MUST 使用`dir=<path>`，`wasm`命令不得读取`p`或`plugin-dir`作为插件选择或路径输入。布尔参数 MUST 只接受`true`、`false`、`1`和`0`；动态`WASM`生成派发器的布尔路由值也 MUST 使用同一标准集合。发布标签校验 MUST 通过显式`tag=<version>`参数传入待校验标签，不得读取环境变量作为隐式标签来源。镜像构建 registry MUST 只来自`hack/config.yaml`或显式`registry=<prefix>`参数，不得读取`LINAPRO_IMAGE_REGISTRY`环境变量。镜像构建内部入口不得保留环境打印调试参数或选项，也不得为这些旧调试入口保留专门判断分支。

#### Scenario: 省略插件模式时自动探测

- **WHEN** 开发者运行`linactl build`且未传入`plugins`
- **THEN** 系统根据`apps/lina-plugins`下是否存在插件清单选择宿主模式或插件完整模式
- **AND** 该自动探测不需要公开`plugins=auto`参数值

#### Scenario: 显式`plugins=auto`被拒绝

- **WHEN** 开发者运行`linactl build plugins=auto`
- **THEN** 系统拒绝执行
- **AND** 错误消息说明`plugins`只接受标准布尔值，自动探测应通过省略该参数触发

#### Scenario: 下划线参数 key 不再映射

- **WHEN** 开发者运行下划线形式的基础镜像参数
- **THEN** 系统不将该参数解释为`base-image`
- **AND** 需要覆盖基础镜像时必须使用`base-image=alpine:3.22`

#### Scenario: Make 包装入口不转发旧参数名

- **WHEN** 开发者通过`make release.tag.check`调用发布校验
- **THEN** 根`Makefile`只转发`print-version`参数
- **AND** 不再转发下划线形式的旧参数名

#### Scenario: `wasm`单插件构建使用`dir`

- **WHEN** 开发者运行`linactl wasm dir=apps/lina-plugins/<plugin-id>`或`make wasm dir=apps/lina-plugins/<plugin-id>`
- **THEN** 系统只从该目录构建单个动态插件产物
- **AND** 相对目录按仓库根目录解析

#### Scenario: `wasm`不读取旧插件选择参数

- **WHEN** 开发者运行`linactl wasm p=<plugin-id>`或`linactl wasm plugin-dir=<path>`
- **THEN** 系统不得将`p`解释为单插件选择
- **AND** 不得将`plugin-dir`解释为显式插件源码目录
- **AND** 根`Makefile`的`wasm`目标不得转发`p`参数

#### Scenario: 宽松布尔别名被拒绝

- **WHEN** 开发者运行`linactl image push=on`
- **THEN** 系统拒绝该布尔值
- **AND** 错误消息说明布尔值只接受`true`、`false`、`1`或`0`

#### Scenario: 动态`WASM`布尔路由值只使用标准集合

- **WHEN** 动态`WASM`生成派发器解析布尔路由值
- **THEN** 系统只将`true`和`1`解析为`true`
- **AND** 其它非标准布尔文本不得作为`true`处理

#### Scenario: 发布标签校验只接受显式标签

- **WHEN** 开发者运行`linactl release.tag.check`
- **AND** 未传入`tag=<version>`
- **THEN** 系统拒绝执行
- **AND** 不从环境变量中读取待校验标签

#### Scenario: 镜像 registry 不读取环境变量

- **WHEN** 环境变量`LINAPRO_IMAGE_REGISTRY`存在
- **AND** 开发者运行`linactl image`且未传入`registry`
- **THEN** 系统不使用该环境变量覆盖镜像 registry
- **AND** registry 仅来自`hack/config.yaml`中的`image.registry`

#### Scenario: 镜像构建环境打印入口不可用

- **WHEN** 调用方传入环境打印调试参数或内部选项
- **THEN** 系统不得进入环境打印流程
- **AND** 不输出构建环境兼容信息
- **AND** 不保留旧调试入口的专门判断分支
