这是无人值守的 `monthly CI` 归档任务。请使用仓库内 `.agents/skills/lina-auto-archive/SKILL.md` 的规则执行自动归档：
1. 只扫描 `openspec/changes/` 根目录下的活跃变更，排除 `archive` 目录。
2. 只归档 `OpenSpec` 状态完成且 `tasks.md` 已全部完成的变更。
3. 使用 `openspec archive -y`，不要使用 `--no-validate`，不要手动移动目录。
4. 不要修改与 `OpenSpec` 归档无关的文件；如果无法安全执行，请失败并说明原因。
