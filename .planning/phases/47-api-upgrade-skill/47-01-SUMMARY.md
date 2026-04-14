---
phase: 47-api-upgrade-skill
plan: "01"
subsystem: tooling
tags: [python, api-upgrade, skill, automation]
dependency_graph:
  requires: []
  provides: [upgrade_version.py]
  affects: [internal/client/client.go, internal/testmock/server.go, internal/testmock/handlers/]
tech_stack:
  added: []
  patterns: [argparse-mutually-exclusive-group, namedtuple-FileChange, pathlib-write_text]
key_files:
  created:
    - .claude/skills/api-upgrade/scripts/upgrade_version.py
  modified: []
decisions:
  - Replace only last occurrence of "OLD" on the versions: line in server.go to preserve older version strings (2.12, 2.15)
  - --dry-run is default; mutual exclusion with --apply enforced via argparse group
  - Missing client.go/server.go/handlers dir are non-fatal warnings to support partial runs
metrics:
  duration_minutes: 5
  completed_date: "2026-04-14"
  tasks_completed: 1
  tasks_total: 1
  files_changed: 1
---

# Phase 47 Plan 01: upgrade_version.py — Mechanical API Version Replacement

**One-liner:** Python 3.10+ stdlib CLI that replaces FlashBlade API version strings in client.go, server.go, and all 38 handler files with --dry-run (default) and --apply modes.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Write upgrade_version.py with dry-run and apply modes | fbe0767 | .claude/skills/api-upgrade/scripts/upgrade_version.py |

## What Was Built

`upgrade_version.py` performs mechanical version-string replacement across three target sets:

1. `internal/client/client.go` — `const APIVersion = "OLD"` → `const APIVersion = "NEW"` via regex
2. `internal/testmock/server.go` — replaces only the last (current) version in the versions slice, preserving 2.12/2.15
3. `internal/testmock/handlers/*.go` (38 files) — `/api/OLD/` → `/api/NEW/` via str.replace per line

**CLI:**
```bash
python3 upgrade_version.py --from 2.22 --to 2.23 [--dry-run | --apply] [--project-root /path]
```

Dry-run (default) prints `file:lineno: old_snippet → new_snippet` for every change. Apply writes files and reports `Done. N files modified.`. Exits 1 if version not found anywhere.

## Decisions Made

- Server.go replacement uses `rfind()` on the versions line to target only the last (current) version string, leaving older ones untouched
- `--dry-run` is the default (mutual exclusion group with `--apply`); no mode flag = dry-run
- Missing target files/dirs produce stderr warnings and are skipped — allows partial runs during incremental upgrades

## Deviations from Plan

None — plan executed exactly as written.

## Verification

```
DRY RUN — no files modified
Would replace '2.22' → '2.23' in 38 file(s):
...client.go:21: const APIVersion = "2.22" → const APIVersion = "2.23"
...server.go:63: "versions": []string{"2.12", "2.15", "2.22"}, → ...{"2.12", "2.15", "2.23"},
...handlers/array_admin.go:62: mux.HandleFunc("/api/2.22/dns", ...) → mux.HandleFunc("/api/2.23/dns", ...)
[35 more handler files]
EXIT: 0
git diff --stat → 0 provider files changed (dry-run confirmed)
```

## Known Stubs

None.

## Self-Check: PASSED

- `.claude/skills/api-upgrade/scripts/upgrade_version.py` — FOUND
- Commit fbe0767 — FOUND
