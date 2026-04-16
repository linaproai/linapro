# @vben/vsh

`vsh` 是一个面向工作区维护和开发自动化的 shell 工具集。

## 特性

- 基于 Node.js 的现代 shell 工具
- 依赖检查与分析
- 循环依赖扫描
- 发布前校验辅助能力

## 安装

```bash
pnpm add -D @vben/vsh
```

## 用法

```bash
vsh [command]
```

常用命令：

- `vsh check-deps`
- `vsh scan-circular`
- `vsh publish-check`
