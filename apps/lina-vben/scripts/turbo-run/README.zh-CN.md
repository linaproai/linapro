# @vben/turbo-run

`turbo-run` 是一个命令行工具，用于在 monorepo 中交互式选择包，并在这些包上并行执行同一个脚本。

## 特性

- 交互式选择目标包
- 面向 monorepo 的脚本发现
- 精确的目标过滤能力

## 安装

```bash
pnpm add -D @vben/turbo-run
```

## 用法

```bash
turbo-run [script]
```

示例：

```bash
turbo-run dev
```
