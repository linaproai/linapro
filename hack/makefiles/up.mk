# Lina Git Target
# ================

## up: 自动生成commit message并推送 [tool|t=claude|codex] [model|m=模型名]
up:
	@set -e; \
	if git diff --quiet HEAD && git diff --cached --quiet && [ -z "$$(git ls-files --others --exclude-standard)" ]; then \
		if git diff --quiet HEAD origin/$$(git branch --show-current) 2>/dev/null; then \
			echo "No changes to commit and nothing to push"; \
		else \
			echo "No local changes, pushing unpushed commits..."; \
			git push origin $$(git branch --show-current); \
		fi; \
		exit 0; \
	fi; \
	git add -A; \
	AI_TOOL="$(strip $(or $(tool),$(t),claude))"; \
	AI_MODEL="$(strip $(or $(model),$(m)))"; \
	if [ "$$AI_TOOL" = "claude" ] && [ -z "$$AI_MODEL" ]; then \
		AI_MODEL="haiku"; \
	fi; \
	if [ "$$AI_TOOL" = "codex" ] && [ -z "$$AI_MODEL" ]; then \
		AI_MODEL="gpt-5.4"; \
	fi; \
	case "$$AI_TOOL" in \
		claude|codex) ;; \
		*) echo "Error: Unsupported AI tool '$$AI_TOOL'. Supported tools: claude, codex"; exit 1 ;; \
	esac; \
	command -v "$$AI_TOOL" >/dev/null 2>&1 || { echo "Error: '$$AI_TOOL' command not found"; exit 1; }; \
	PROMPT="Analyze the git diff above and generate a concise commit message (single line, max 72 chars, lowercase, no quotes). Output only the commit message itself, nothing else."; \
	generate_input() { \
		git diff --cached --stat; \
		echo "---"; \
		git diff --cached | head -2000; \
	}; \
	echo "Analyzing changes and generating commit message via AI (tool: $$AI_TOOL, model: $${AI_MODEL:-default})..."; \
	case "$$AI_TOOL" in \
		claude) \
			if [ -n "$$AI_MODEL" ]; then \
				MSG=$$(generate_input | claude -p "$$PROMPT" --model "$$AI_MODEL") || { echo "Error: Claude command failed"; exit 1; }; \
			else \
				MSG=$$(generate_input | claude -p "$$PROMPT") || { echo "Error: Claude command failed"; exit 1; }; \
			fi; \
			;; \
		codex) \
			TMP_FILE=$$(mktemp); \
			ERR_FILE=$$(mktemp); \
			trap 'rm -f "$$TMP_FILE" "$$ERR_FILE"' EXIT; \
			if [ -n "$$AI_MODEL" ]; then \
				generate_input | codex exec --sandbox read-only --color never --model "$$AI_MODEL" -o "$$TMP_FILE" "$$PROMPT" >/dev/null 2>"$$ERR_FILE" || { echo "Error: Codex command failed"; tail -20 "$$ERR_FILE"; rm -f "$$TMP_FILE" "$$ERR_FILE"; exit 1; }; \
			else \
				generate_input | codex exec --sandbox read-only --color never -o "$$TMP_FILE" "$$PROMPT" >/dev/null 2>"$$ERR_FILE" || { echo "Error: Codex command failed"; tail -20 "$$ERR_FILE"; rm -f "$$TMP_FILE" "$$ERR_FILE"; exit 1; }; \
			fi; \
			MSG=$$(cat "$$TMP_FILE"); \
			rm -f "$$TMP_FILE" "$$ERR_FILE"; \
			trap - EXIT; \
			;; \
	esac; \
	COMMIT_MSG=$$(echo "$$MSG" | tail -1); \
	if [ -z "$$COMMIT_MSG" ]; then \
		echo "Error: Failed to generate commit message"; \
		exit 1; \
	fi; \
	echo "Commit: $$COMMIT_MSG"; \
	git commit -m "$$COMMIT_MSG" && \
	git push origin $$(git branch --show-current)
