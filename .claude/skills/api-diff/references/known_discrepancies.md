# Known Discrepancies — FlashBlade Swagger vs Real API

This file tracks confirmed divergences between the swagger spec and actual FlashBlade API behavior.
Update this file when a `needs_verification` diff item is investigated and resolved.

## How to Use

1. Run `diff_swagger.py` — new items default to `annotation: needs_verification`
2. Investigate the item (test against a real array or check release notes)
3. Add an entry below with the confirmed annotation
4. Re-run `diff_swagger.py --discrepancies .claude/skills/api-diff/references/known_discrepancies.json`
   to apply overrides to the diff output

## Discrepancy Annotations

| Value | Meaning |
|-------|---------|
| `real_change` | Confirmed API behavior change between versions |
| `swagger_artifact` | Present in swagger but not in real API, or swagger doc error |
| `needs_verification` | Default — not yet investigated |

## Confirmed Discrepancies

<!-- Add entries below as they are discovered -->
<!-- Format: version range, path, method, annotation, note -->

| Version Range | Normalized Path | Method | Annotation | Notes |
|---------------|-----------------|--------|------------|-------|
| (none yet) | | | | |

## Investigation Log

<!-- Free-form notes per investigation session -->
