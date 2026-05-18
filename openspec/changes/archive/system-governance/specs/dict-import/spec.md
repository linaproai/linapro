## ADDED Requirements

### Requirement: 字典类型导入
系统 SHALL 支持从 Excel 文件导入字典类型记录。系统 SHALL 提供模板下载端点，并在持久化前校验导入数据。

#### Scenario: 下载字典类型导入模板
- **WHEN** 用户请求 `GET /api/v1/dict/type/import-template`
- **THEN** 系统返回包含示例数据的 Excel 模板，列包含：字典名称、字典类型、状态、备注

#### Scenario: 导入有效字典类型数据
- **WHEN** 用户上传有效的 Excel 文件到 `POST /api/v1/dict/type/import`
- **THEN** 系统校验所有行，创建记录，返回成功数量

#### Scenario: 导入字典类型数据校验失败
- **WHEN** 用户上传包含无效数据的 Excel 文件（缺少必填字段、字典类型值重复）
- **THEN** 系统拒绝整个导入，返回包含行号和原因的错误详情

#### Scenario: 覆盖模式导入字典类型
- **WHEN** 用户上传 Excel 文件时设置 `updateSupport=true`，文件中包含已存在的字典类型值
- **THEN** 系统使用导入的值更新已有记录

#### Scenario: 忽略模式导入字典类型
- **WHEN** 用户上传 Excel 文件时设置 `updateSupport=false`（默认），文件中包含已存在的字典类型值
- **THEN** 系统跳过已存在的记录，仅创建新记录

#### Scenario: 字典类型导入弹窗 UI
- **WHEN** 用户点击字典类型管理页面的"导入"按钮
- **THEN** 系统展示弹窗，包含模板下载链接、拖拽上传区域、文件类型提示（xlsx/xls）、覆盖/忽略模式开关

### Requirement: 字典数据导入
系统 SHALL 支持从 Excel 文件导入字典数据记录。系统 SHALL 提供模板下载端点，并在持久化前校验导入数据。

#### Scenario: 下载字典数据导入模板
- **WHEN** 用户请求 `GET /api/v1/dict/data/import-template`
- **THEN** 系统返回包含示例数据的 Excel 模板，列包含：字典类型、字典标签、字典键值、排序、Tag样式、CSS类名、状态、备注

#### Scenario: 导入有效字典数据
- **WHEN** 用户上传有效的 Excel 文件到 `POST /api/v1/dict/data/import`
- **THEN** 系统校验所有行，创建记录，返回成功数量

#### Scenario: 导入字典数据校验失败
- **WHEN** 用户上传包含无效数据的 Excel 文件（缺少必填字段、字典类型不存在）
- **THEN** 系统拒绝整个导入，返回包含行号和原因的错误详情

#### Scenario: 覆盖模式导入字典数据
- **WHEN** 用户上传 Excel 文件时设置 `updateSupport=true`，文件中包含匹配 dictType+value 组合的记录
- **THEN** 系统使用导入的值更新已有记录

#### Scenario: 忽略模式导入字典数据
- **WHEN** 用户上传 Excel 文件时设置 `updateSupport=false`（默认），文件中包含匹配 dictType+value 组合的记录
- **THEN** 系统跳过已存在的记录，仅创建新记录

#### Scenario: 字典数据导入弹窗 UI
- **WHEN** 用户点击字典数据管理页面的"导入"按钮
- **THEN** 系统展示弹窗，包含模板下载链接、拖拽上传区域、文件类型提示（xlsx/xls）、覆盖/忽略模式开关

### Requirement: 字典合并导出
系统 SHALL 提供 `GET /api/v1/dict/export` 合并导出接口，同时导出字典类型和字典数据到双 Sheet Excel 文件。

#### Scenario: 合并导出字典
- **WHEN** 用户请求 `GET /api/v1/dict/export`
- **THEN** 系统生成包含两个 Sheet 的 Excel 文件：第一个 Sheet 为字典类型数据，第二个 Sheet 为字典数据

### Requirement: 字典合并导入
系统 SHALL 提供 `POST /api/v1/dict/import` 合并导入接口，支持同时导入字典类型和字典数据。

#### Scenario: 合并导入字典
- **WHEN** 用户上传包含两个 Sheet 的 Excel 文件到 `POST /api/v1/dict/import`
- **THEN** 系统从第一个 Sheet 读取字典类型数据，从第二个 Sheet 读取字典数据，分别校验并导入，返回成功数量

#### Scenario: 覆盖模式合并导入
- **WHEN** 用户上传 Excel 文件时设置 `updateSupport=true`
- **THEN** 系统对已存在的字典类型和字典数据执行更新操作

### Requirement: 字典导入模板下载
系统 SHALL 提供字典导入模板下载接口，返回包含两个 Sheet 的模板文件。

#### Scenario: 下载合并导入模板
- **WHEN** 用户请求 `GET /api/v1/dict/import-template`
- **THEN** 系统返回包含两个 Sheet 的 Excel 模板：字典类型模板和字典数据模板

### Requirement: 字典类型级联删除
系统 SHALL 在删除字典类型时同时删除关联的字典数据。

#### Scenario: 删除字典类型
- **WHEN** 用户删除一个字典类型
- **THEN** 系统同时删除该字典类型下所有关联的字典数据记录

#### Scenario: 删除确认提示
- **WHEN** 用户点击字典类型的"删除"按钮
- **THEN** 确认弹窗提示用户将同时删除关联的字典数据

### Requirement: 前端字典管理导出导入优化

#### Scenario: 字典类型面板使用合并接口
- **WHEN** 用户在字典类型面板点击"导出"或"导入"
- **THEN** 使用合并导出导入接口，同时处理字典类型和字典数据

#### Scenario: 字典数据面板移除导出导入按钮
- **WHEN** 用户访问字典数据面板
- **THEN** 不再显示独立的"导出"和"导入"按钮
