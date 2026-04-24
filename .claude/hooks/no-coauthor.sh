#!/usr/bin/env bash
# PreToolUse hook: block git commit commands containing Co-Authored-By trailer.
# Project policy (CLAUDE.md + CONVENTIONS.md) strictly forbids co-author lines.
# This catches Claude Code sub-agents that include the trailer before --no-verify
# is applied, providing a second layer of defense alongside the commit-msg git hook.
#
# Exit codes (Claude Code PreToolUse semantics):
#   0  → allow
#   2  → BLOCK (hard rejection, tool call not executed)
set -euo pipefail

input=$(cat)
tool=$(echo "$input" | jq -r '.tool_name // ""')

[[ "$tool" != "Bash" ]] && exit 0

cmd=$(echo "$input" | jq -r '.tool_input.command // ""')

# Only inspect git commit invocations
if ! echo "$cmd" | grep -qE '(^|[[:space:]])git[[:space:]].*commit'; then
  exit 0
fi

# Check for Co-Authored-By in the command string (covers -m "...", HEREDOC, etc.)
if echo "$cmd" | grep -iE 'co-?authored-?by' > /dev/null 2>&1; then
  cat >&2 <<'EOF'
[no-coauthor] BLOCKED: git commit contains a Co-Authored-By trailer.
Project policy (CLAUDE.md + CONVENTIONS.md) strictly forbids co-author lines.
Remove the trailer from the commit message and retry.
EOF
  exit 2
fi

exit 0
