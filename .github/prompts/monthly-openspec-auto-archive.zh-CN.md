这是无人值守的 `monthly CI` 自动归档任务。请使用仓库内 `.agents/skills/lina-openspec-archive-changes/SKILL.md` 的规则执行自动归档：
1. 只扫描 `openspec/changes/` 根目录下的活跃变更，明确排除 `openspec/changes/archive/`。
2. 只归档已经完成的变更；如状态、任务清单或 OpenSpec 输出无法安全判定为完成，请失败并说明原因。
3. 在执行任何归档命令前，先对所有候选完成变更做完整预检：`openspec status --change "<change-name>" --json` 的 artifact 必须全部完成，`tasks.md` 不得缺失且不得含未完成任务，所有 `MODIFIED`/`REMOVED` requirement header 必须能在当前 `openspec/specs/<capability>/spec.md` 中找到。
4. 如果任一候选未通过预检，必须失败并输出候选名与原因，不得先归档其他候选。
5. 对每个可归档变更使用 OpenSpec CLI 执行 `openspec archive -y "<change-name>"`，不要手动移动目录绕过 OpenSpec。
6. 归档后重新确认对应变更已经离开活跃变更列表。
7. 归档变更范围必须保持在 `openspec/**` 内。
8. 不要等待人工交互；如果无法在 `CI` 中安全完成，请失败并说明原因。
