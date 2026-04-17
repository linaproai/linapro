# Proposal: redirect-dev-logs-to-project-temp

当前仓库根 `make dev` 会把前端与后端开发阶段的标准输出统一重定向到 `/tmp/lina-core.log` 与 `/tmp/lina-vben.log`。这会让开发者在排查问题时需要跳出项目目录查找日志文件，也不利于将开发期临时产物统一收敛到仓库根 `temp/` 目录。本次反馈变更将开发日志输出路径调整为项目根目录 `temp/`，并同步更新相关状态提示，确保日常开发测试时能在仓库内直接查看日志定位问题。
