#!/bin/bash
headroom proxy --port 8787 &
HEADROOM_PID=$!
trap "kill $HEADROOM_PID 2>/dev/null" EXIT INT TERM

CLAUDE_CODE_DISABLE_1M_CONTEXT=1 ANTHROPIC_BASE_URL=http://127.0.0.1:8787 claude --allow-dangerously-skip-permissions
