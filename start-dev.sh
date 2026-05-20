#!/bin/bash
headroom proxy --port 8787 &
HEADROOM_PID=$!
trap "kill $HEADROOM_PID 2>/dev/null" EXIT INT TERM

ANTHROPIC_BASE_URL=http://127.0.0.1:8787 claude --allow-dangerously-skip-permissions
