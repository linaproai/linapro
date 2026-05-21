这是无人值守的 `monthly CI` 自动归档任务。请使用仓库内 `.agents/skills/lina-auto-archive/SKILL.md` 的规则执行自动归档：
1. 只扫描 `openspec/changes/` 根目录下的活跃变更，明确排除 `openspec/changes/archive/`。
2. 只归档已经完成的变更；如状态、任务清单或 OpenSpec 输出无法安全判定为完成，请失败并说明原因。
3. 对每个可归档变更使用 OpenSpec CLI 执行 `openspec archive -y "<change-name>"`，不要手动移动目录绕过 OpenSpec。
4. 归档后重新确认对应变更已经离开活跃变更列表。
5. 归档变更范围必须保持在 `openspec/**` 内。
6. 不要等待人工交互；如果无法在 `CI` 中安全完成，请失败并说明原因。
