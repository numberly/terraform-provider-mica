#!/usr/bin/env bash
# PreToolUse hook:
#   - BLOCK Grep/Glob targeting .go / .tf files (exit 2)
#   - WARN on Bash rg/grep/ag/ack targeting .go / .tf files (exit 1, non-blocking)
# Forces Serena MCP usage for code navigation on Go and Terraform sources.
# Read is intentionally NOT intercepted (required before Edit).
set -euo pipefail

input=$(cat)
tool=$(echo "$input" | jq -r '.tool_name // ""')

if [[ "$tool" == "Bash" ]]; then
  cmd=$(echo "$input" | jq -r '.tool_input.command // ""')
  # Match rg/grep/ag/ack invocations referencing .go or .tf (as arg, glob, or --type)
  if [[ "$cmd" =~ (^|[^[:alnum:]_])(rg|grep|ag|ack)([[:space:]]|$) ]]; then
    if [[ "$cmd" =~ \.(go|tf)([[:space:]\"\'\)]|$) ]] \
       || [[ "$cmd" =~ --type[=[:space:]]+(go|terraform|tf|hcl) ]] \
       || [[ "$cmd" =~ -t[[:space:]]+(go|terraform|tf|hcl) ]]; then
      cat >&2 <<'EOF'
[serena-first] WARNING: rg/grep on .go/.tf detected. Prefer Serena MCP:
  - mcp__serena__find_symbol
  - mcp__serena__get_symbols_overview
  - mcp__serena__search_for_pattern (with relative_path)
Proceeding anyway (non-blocking).
EOF
      exit 1
    fi
  fi
  exit 0
fi

if [[ "$tool" != "Grep" && "$tool" != "Glob" ]]; then
  exit 0
fi

type=$(echo "$input"    | jq -r '.tool_input.type    // ""')
glob=$(echo "$input"    | jq -r '.tool_input.glob    // ""')
path=$(echo "$input"    | jq -r '.tool_input.path    // ""')
pattern=$(echo "$input" | jq -r '.tool_input.pattern // ""')

hit=0
case "$type" in
  go|terraform|tf|hcl) hit=1 ;;
esac

[[ "$glob" == *".go"* || "$glob" == *".tf"* ]] && hit=1
[[ "$path" == *.go || "$path" == *.tf ]] && hit=1

# Glob tool: the pattern field IS the glob. Grep pattern is regex, skip it.
if [[ "$tool" == "Glob" ]]; then
  [[ "$pattern" == *".go"* || "$pattern" == *".tf"* ]] && hit=1
fi

if [[ $hit -eq 1 ]]; then
  cat >&2 <<'EOF'
[serena-first] Grep/Glob blocked on .go/.tf targets in this project.
Use Serena MCP instead:
  - Symbol / definition → mcp__serena__find_symbol
  - References          → mcp__serena__find_referencing_symbols
  - File overview       → mcp__serena__get_symbols_overview
  - Pattern in symbol   → mcp__serena__search_for_pattern (with relative_path)
Read remains allowed (required before Edit).
If you truly need a raw text search, use `rg` via Bash.
EOF
  exit 2
fi

exit 0
