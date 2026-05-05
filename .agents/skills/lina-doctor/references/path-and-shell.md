# PATH 与 Shell 处理

`lina-doctor`只在当前进程中临时修复`PATH`，并打印用户可复制的持久化命令。脚本不得自动写入用户的 rc 文件或 PowerShell`$PROFILE`。

## GoFrame CLI

`go install github.com/gogf/gf/v2/cmd/gf@latest`默认把二进制写入`$GOBIN`。如果`GOBIN`为空，Go 默认使用`$HOME/go/bin`。

当前 shell 临时修复：

```bash
export PATH="$HOME/go/bin:$PATH"
```

持久化示例：

```bash
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bashrc
fish_add_path "$HOME/go/bin"
```

## npm 全局 Prefix

`pnpm`和`@fission-ai/openspec`通过 npm 全局安装时，二进制通常位于：

```bash
$(npm config get prefix)/bin
```

当前 shell 临时修复：

```bash
export PATH="$(npm config get prefix)/bin:$PATH"
```

如果 npm 全局 prefix 指向系统目录并触发`EACCES`，优先把 npm prefix 调整到用户目录，而不是默认使用 root 权限。

## PowerShell

PowerShell 用户需要根据当前主机版本确认`$PROFILE`路径：

```powershell
$PROFILE
```

持久化 PATH 时应使用用户级路径设置或手动编辑`$PROFILE`。`lina-doctor`不直接写入这些文件。

## 冲突处理

如果用户已有 rc 文件包含多段 PATH 逻辑，`lina-doctor`只打印建议，不删除、不排序、不覆盖既有内容。用户可手动把建议命令追加到最靠近文件末尾的位置，避免被后续 PATH 重置覆盖。
