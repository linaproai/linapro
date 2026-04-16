## 技术方案

### 1. 数据库设计

新增 `sys_file` 文件管理表：

```sql
CREATE TABLE sys_file (
  id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '文件ID',
  name       VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '存储文件名',
  original   VARCHAR(255)    NOT NULL DEFAULT '' COMMENT '原始文件名',
  suffix     VARCHAR(32)     NOT NULL DEFAULT '' COMMENT '文件后缀',
  size       BIGINT UNSIGNED NOT NULL DEFAULT 0  COMMENT '文件大小（字节）',
  hash       VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '文件SHA-256散列值，用于去重',
  url        VARCHAR(512)    NOT NULL DEFAULT '' COMMENT '文件访问URL',
  path       VARCHAR(512)    NOT NULL DEFAULT '' COMMENT '文件存储路径',
  engine     VARCHAR(32)     NOT NULL DEFAULT 'local' COMMENT '存储引擎：local=本地',
  created_by BIGINT UNSIGNED NOT NULL DEFAULT 0  COMMENT '上传者用户ID',
  created_at DATETIME        NOT NULL COMMENT '上传时间',
  updated_at DATETIME        NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  INDEX idx_engine (engine),
  INDEX idx_created_by (created_by),
  INDEX idx_suffix (suffix),
  INDEX idx_hash (hash)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件管理';
```

使用场景直接记录在 `sys_file` 表的 `scene` 字段中。相同散列值的文件如在不同场景使用，则各自创建独立的文件记录。

### 2. 后端架构

#### 2.1 文件存储抽象层

在 `internal/service/file/` 中设计存储接口，默认实现本地存储，后续可扩展 OSS：

```
service/file/
├── file.go          # 主服务：上传/下载/列表/删除业务逻辑
├── storage.go       # Storage 接口定义
└── storage_local.go # 本地存储实现
```

**Storage 接口**：
- `Put(ctx, filename, data) (path, error)` — 保存文件
- `Get(ctx, path) (data, error)` — 读取文件
- `Delete(ctx, path) error` — 删除文件
- `Url(ctx, path) string` — 获取访问 URL

#### 2.2 API 端点设计

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/file/upload` | 上传文件（multipart/form-data） |
| GET | `/file` | 文件列表（分页+筛选） |
| GET | `/file/download/{id}` | 下载文件 |
| DELETE | `/file/{ids}` | 删除文件（支持批量） |

上传接口返回：
```json
{
  "id": 1,
  "name": "20260319_abc123.png",
  "original": "avatar.png",
  "url": "/api/v1/uploads/2026/03/20260319_abc123.png",
  "suffix": "png",
  "size": 102400
}
```

### 3. 前端架构

#### 3.1 通用上传组件

参考 ruoyi-plus-vben5 的实现，在 `src/components/upload/` 创建：

- `FileUpload` — 文件上传组件（支持拖拽上传）
- `ImageUpload` — 图片上传组件（picture-card 样式）
- 通用上传 hook（`useUpload`）处理进度、校验、错误等

组件通过 `v-model:value` 绑定文件 ID（单文件为 string，多文件为 string[]），回显时通过文件 ID 查询文件信息。

#### 3.2 文件管理页面

在系统管理菜单下新增"文件管理"子菜单，页面功能包含：
- 文件列表表格（文件名、原始名、后缀、预览、上传时间、上传者、操作）
- 搜索条件（文件名、后缀、上传时间范围）
- 工具栏按钮（批量删除、文件上传、图片上传）
- 操作列（下载、删除）
- 图片预览切换开关

#### 3.3 现有功能改造

**通知公告编辑器**：
- 给 TiptapEditor 传入 `uploadHandler` prop，调用文件上传接口
- 上传成功后返回 URL 插入编辑器
- 移除 Base64 内嵌模式依赖

**用户头像上传**：
- 改为调用通用文件上传接口 `POST /file/upload`
- 上传后使用返回的 URL 更新用户头像
- 删除原有的 `POST /user/profile/avatar` 独立上传端点及静态文件服务路由

### 4. 文件存储路径

本地存储按年月组织目录结构：
```
temp/upload/
  2026/
    03/
      20260319_abc123.png
      20260319_def456.docx
```

### 5. 配置变更

`config.yaml` 中 `upload.path` 配置项保持不变（`temp/upload`），新增 `upload.maxSize` 配置项控制最大上传文件大小。
