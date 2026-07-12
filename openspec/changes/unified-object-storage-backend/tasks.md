## 1. 规范与宿主后端

- [x] 1.1 撰写 proposal/design/spec
- [x] 1.2 local storage provider 支持 `files/` 前缀路由到 NamespaceFiles
- [x] 1.3 文件中心 FileObjectStore：ResolveProvider + files/ 键 + 本地回退读

## 2. 文件中心接入

- [x] 2.1 Upload/CreateFromReader Put 与 cleanup Delete 走 FileObjectStore
- [x] 2.2 OpenByID/OpenByPath Get 走 FileObjectStore
- [x] 2.3 Delete 物理删除走 FileObjectStore
- [x] 2.4 写入 engine=providerID；DI 接线

## 3. 验证

- [x] 3.1 单测：local 路由、files 前缀、冲突错误映射
- [x] 3.2 Go 编译相关包；i18n：新增 FILE_STORAGE_CONFLICT 英文源码；无缓存新语义；数据权限仍基于 sys_file
