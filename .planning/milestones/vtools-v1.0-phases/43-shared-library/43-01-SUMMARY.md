---
phase: 43-shared-library
plan: 01
subsystem: infra
tags: [python, swagger, allof-resolution, path-normalization, pytest, stdlib]

requires: []
provides:
  - resolve_all_of: merges allOf chains in swagger-2.22.json, 318 schemas resolved, no allOf/$ref at top level
  - normalize_path: strips /api/<version>/ and /oauth2/<version>/ prefixes from swagger paths
  - flatten_schema: merges allOf/$ref into properties/required/description/type dict
  - pytest.ini with pythonpath=.claude/skills for _shared imports

affects:
  - 44-swagger-to-reference
  - 45-api-diff
  - 46-api-upgrade

tech-stack:
  added: [pytest-ini-pythonpath, python-stdlib-only]
  patterns:
    - stdlib-only Python skill utilities (no external dependencies)
    - PYTHONPATH=.claude/skills for _shared package resolution
    - pytest.ini pythonpath for zero-setup test discovery from project root

key-files:
  created:
    - .claude/skills/_shared/__init__.py
    - .claude/skills/_shared/swagger_utils.py
    - .claude/skills/_shared/test_swagger_utils.py
    - .claude/skills/conftest.py
    - conftest.py
    - pytest.ini
  modified: []

key-decisions:
  - "PYTHONPATH=.claude/skills required for standalone python3 -c imports; pytest.ini pythonpath handles pytest discovery automatically"
  - "resolve_all_of skips *Response/*GetResponse wrappers and _-prefixed private schemas from standalone output"
  - "conftest.py at .claude/skills/ and root for sys.path setup; pytest.ini pythonpath= as primary mechanism"

patterns-established:
  - "Shared Python utilities live in .claude/skills/_shared/swagger_utils.py, importable as 'from _shared.swagger_utils import ...'"
  - "All skill Python scripts use stdlib only — no pip install required"
  - "Tests run from project root: python3 -m pytest .claude/skills/_shared/test_swagger_utils.py -v"

requirements-completed:
  - SLIB-01
  - SLIB-02

duration: 10min
completed: 2026-04-14
---

# Phase 43 Plan 01: Shared Library Summary

**stdlib-only swagger_utils.py with resolve_all_of (318 schemas, no allOf/$ref), normalize_path (strips /api/<version>/ prefixes), flatten_schema — 15 pytest tests all passing**

## Performance

- **Duration:** 10 min
- **Started:** 2026-04-14T07:51:25Z
- **Completed:** 2026-04-14T08:01:08Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- resolve_all_of resolves 318 schemas from swagger-2.22.json with no remaining allOf or $ref at top level
- normalize_path handles all path formats: /api/<version>/, /api/ (no version), /oauth2/<version>/
- flatten_schema merges allOf/$ref chains into unified properties/required/description/type dict
- 15 pytest tests covering all three functions pass in 0.09s from project root

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement swagger_utils.py** - `5da1cfc` (feat)
2. **Task 2: Write test suite and validate against swagger-2.22.json** - `f611a1a` (test)

**Plan metadata:** (final commit below)

## Files Created/Modified

- `.claude/skills/_shared/__init__.py` - Package marker for _shared Python package
- `.claude/skills/_shared/swagger_utils.py` - resolve_all_of, normalize_path, flatten_schema (stdlib only)
- `.claude/skills/_shared/test_swagger_utils.py` - 15 pytest tests for all three functions
- `.claude/skills/conftest.py` - sys.path setup for _shared imports
- `conftest.py` - Root conftest for pytest path setup
- `pytest.ini` - pythonpath=.claude/skills for zero-setup test discovery

## Decisions Made

- PYTHONPATH=.claude/skills is required for standalone `python3 -c` imports from project root; `pytest.ini` pythonpath= handles pytest automatically
- resolve_all_of skips *Response/*GetResponse wrappers and _-prefixed private schemas from the output dict (resolved inline when referenced, not emitted standalone)
- Max recursion depth guard at 20 prevents circular reference infinite loops

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added pytest.ini and conftest.py for _shared import resolution**

- **Found during:** Task 1 (verify import from project root)
- **Issue:** `from _shared.swagger_utils import ...` fails from project root because .claude/skills/ is not on sys.path
- **Fix:** Added pytest.ini with `pythonpath = .claude/skills` (handles pytest), conftest.py at .claude/skills/ and project root (handles sys.path for direct invocation); PYTHONPATH=.claude/skills needed for standalone python3 -c
- **Files modified:** pytest.ini, conftest.py, .claude/skills/conftest.py
- **Verification:** `python3 -m pytest .claude/skills/_shared/test_swagger_utils.py -v` passes 15/15; PYTHONPATH=.claude/skills python3 -c import succeeds
- **Committed in:** 5da1cfc (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (blocking import resolution)
**Impact on plan:** Necessary for pytest test discovery to work without environment setup. Standalone python3 -c commands require PYTHONPATH=.claude/skills.

## Issues Encountered

None beyond the import path resolution deviation above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- _shared library is complete and tested; ready for Phase 44 (swagger-to-reference) and Phase 45 (api-diff)
- Downstream skills import via: `from _shared.swagger_utils import resolve_all_of, normalize_path, flatten_schema`
- Run tests with: `python3 -m pytest .claude/skills/_shared/test_swagger_utils.py -v` from project root
- For standalone scripts: set `PYTHONPATH=.claude/skills` or use `sys.path.insert(0, '.claude/skills')`

---
*Phase: 43-shared-library*
*Completed: 2026-04-14*
