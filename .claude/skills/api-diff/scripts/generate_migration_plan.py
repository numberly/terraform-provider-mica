"""
generate_migration_plan.py — Turn a diff.json into an actionable migration plan.

Usage:
  python3 generate_migration_plan.py <diff.json> <ROADMAP.md> [--output path] [--format json|markdown]

Output categories:
  update_models  — modified schemas that need Go struct updates
  new_resources  — new endpoints to evaluate as Terraform resources
  deprecated     — removed endpoints/schemas
  roadmap_gaps   — new endpoints that match ROADMAP.md Candidate/Deferred entries

stdlib only: json, argparse, re, sys, os, pathlib
"""
from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Any


# ---------------------------------------------------------------------------
# ROADMAP parsing
# ---------------------------------------------------------------------------

def _parse_roadmap_not_implemented(roadmap_path: str) -> list[dict[str, str]]:
    """
    Parse the '## Not Implemented' section of ROADMAP.md.

    Returns a list of dicts with keys:
      - api_section: raw name from column 1
      - status: 'Candidate' | 'Deferred'
      - slug: normalized slug for matching (lowercase, words only)
    """
    text = Path(roadmap_path).read_text(encoding="utf-8")

    # Find the "## Not Implemented" section — everything after that heading
    # until next "## " heading (or end of file)
    match = re.search(r"^## Not Implemented\b", text, re.MULTILINE)
    if not match:
        return []

    section_start = match.end()
    next_h2 = re.search(r"^## ", text[section_start:], re.MULTILINE)
    section_text = text[section_start: section_start + next_h2.start()] if next_h2 else text[section_start:]

    entries: list[dict[str, str]] = []
    for line in section_text.splitlines():
        # Table rows: start with | and are not header (contains ---) or separator
        if not line.startswith("|"):
            continue
        if re.search(r"\|[-: ]+\|", line):
            continue  # separator row

        cols = [c.strip() for c in line.strip("|").split("|")]
        if len(cols) < 2:
            continue

        api_section = cols[0].strip()
        if not api_section or api_section.lower() in ("api section", "api_section"):
            continue  # header row

        # Last column is status
        status = cols[-1].strip()
        if status not in ("Candidate", "Deferred"):
            continue

        entries.append({
            "api_section": api_section,
            "status": status,
            "slug": _make_slug(api_section),
        })

    return entries


def _parse_roadmap_implemented(roadmap_path: str) -> list[dict[str, str]]:
    """
    Parse the '## Implemented' section of ROADMAP.md (all subsections until '## Not Implemented').

    Returns a list of dicts with keys:
      - api_section: raw name from column 1 (e.g. "File Systems")
      - resource_name: Terraform resource name from column 2 (e.g. "flashblade_file_system")
      - slug: normalized slug for matching
    """
    text = Path(roadmap_path).read_text(encoding="utf-8")

    match = re.search(r"^## Implemented\b", text, re.MULTILINE)
    if not match:
        return []

    section_start = match.end()
    # Section ends at "## Not Implemented" or next non-subsection H2
    next_h2 = re.search(r"^## (?!#)", text[section_start:], re.MULTILINE)
    section_text = text[section_start: section_start + next_h2.start()] if next_h2 else text[section_start:]

    entries: list[dict[str, str]] = []
    for line in section_text.splitlines():
        if not line.startswith("|"):
            continue
        if re.search(r"\|[-: ]+\|", line):
            continue  # separator row

        cols = [c.strip() for c in line.strip("|").split("|")]
        if len(cols) < 4:
            continue

        api_section = cols[0].strip()
        if not api_section or api_section.lower() in ("api section", "api_section"):
            continue  # header row

        # Extract resource name — strip backticks, handle "Yes + Yes" style entries
        resource_raw = cols[1].strip().strip("`")
        resource_name = resource_raw if resource_raw.startswith("flashblade_") else None

        # Only include Done entries
        status_col = cols[3].strip() if len(cols) > 3 else ""
        if status_col != "Done":
            continue

        entries.append({
            "api_section": api_section,
            "resource_name": resource_name,
            "slug": _make_slug(api_section),
        })

    return entries


def _make_slug(name: str) -> str:
    """Lowercase slug: letters and digits only, spaces → hyphens."""
    return re.sub(r"[^a-z0-9]+", "-", name.lower()).strip("-")


def _roadmap_words(entry: dict[str, str]) -> list[str]:
    """Extract meaningful words from a roadmap api_section (≥3 chars)."""
    return [w for w in re.split(r"[^a-z0-9]+", entry["api_section"].lower()) if len(w) >= 3]


def _match_roadmap(normalized_path: str, roadmap_entries: list[dict[str, str]]) -> dict[str, str] | None:
    """
    Fuzzy match: return first roadmap entry where ≥2 words from api_section
    appear in normalized_path (case-insensitive).
    """
    path_lower = normalized_path.lower()
    for entry in roadmap_entries:
        words = _roadmap_words(entry)
        if len(words) < 2:
            # Single-word entry: require exact substring match
            if words and words[0] in path_lower:
                return entry
            continue
        matches = sum(1 for w in words if w in path_lower)
        if matches >= 2:
            return entry
    return None


# ---------------------------------------------------------------------------
# Migration plan builder
# ---------------------------------------------------------------------------

_SCHEMA_SUFFIXES = ("Post", "Patch", "Get")


def _schema_base_name(schema_name: str) -> str:
    """Strip Post/Patch/Get suffix to get the base resource schema name."""
    for suffix in _SCHEMA_SUFFIXES:
        if schema_name.endswith(suffix) and len(schema_name) > len(suffix):
            return schema_name[: -len(suffix)]
    return schema_name


def _group_modified_schemas(items: list[dict[str, Any]]) -> list[dict[str, Any]]:
    """
    Group modified schemas by base name (FileSystem + FileSystemPost + FileSystemPatch → FileSystem).
    Merge added/removed/changed fields (union, deduplicated).
    """
    groups: dict[str, dict[str, Any]] = {}
    for item in items:
        base = _schema_base_name(item["schema_name"])
        details = item.get("details", {})
        if base not in groups:
            groups[base] = {
                "schema_name": base,
                "variants": [item["schema_name"]],
                "added_fields": list(details.get("added_fields", [])),
                "removed_fields": list(details.get("removed_fields", [])),
                "changed_fields": list(details.get("changed_fields", [])),
                "annotation": item.get("annotation", "needs_verification"),
            }
        else:
            g = groups[base]
            g["variants"].append(item["schema_name"])
            for f in details.get("added_fields", []):
                if f not in g["added_fields"]:
                    g["added_fields"].append(f)
            for f in details.get("removed_fields", []):
                if f not in g["removed_fields"]:
                    g["removed_fields"].append(f)
            for f in details.get("changed_fields", []):
                if f not in g["changed_fields"]:
                    g["changed_fields"].append(f)
            # Promote annotation: real_change > needs_verification > swagger_artifact
            if item.get("annotation") == "real_change":
                g["annotation"] = "real_change"
    return list(groups.values())


def _match_implemented(
    schema_base: str,
    implemented_entries: list[dict[str, str]],
) -> dict[str, str] | None:
    """
    Match a schema base name (e.g. "FileSystem", "QosPolicy") against implemented
    ROADMAP entries by converting both to slugs and checking for overlap.
    """
    # Convert CamelCase to slug: "FileSystem" → "file-system", "QosPolicy" → "qos-policy"
    slug = re.sub(r"(?<=[a-z0-9])(?=[A-Z])", "-", schema_base).lower()
    slug = re.sub(r"[^a-z0-9]+", "-", slug).strip("-")

    for entry in implemented_entries:
        entry_slug = entry["slug"]
        # Direct slug containment (both directions)
        if slug in entry_slug or entry_slug in slug:
            return entry
        # Word overlap: ≥2 shared words
        slug_words = set(slug.split("-"))
        entry_words = set(entry_slug.split("-"))
        if len(slug_words & entry_words) >= 2:
            return entry
    return None


def _action_for_modified_schema(item: dict[str, Any]) -> str:
    details = item.get("details", {})
    added = details.get("added_fields", [])
    schema = item.get("schema_name", "Unknown")
    if added:
        fields_str = ", ".join(added)
        return f"Add {fields_str} to {schema}Post/Patch structs"
    changed = details.get("changed_fields", [])
    if changed:
        names = ", ".join(f["field"] for f in changed)
        return f"Update field types for {names} in {schema} structs"
    removed = details.get("removed_fields", [])
    if removed:
        fields_str = ", ".join(removed)
        return f"Remove {fields_str} from {schema} structs (check usage)"
    return f"Review {schema} struct for schema changes"


def _action_for_new_resource(normalized_path: str) -> str:
    snake = normalized_path.replace("-", "_").replace("/", "_")
    return f"Evaluate for new Terraform resource flashblade_{snake}"


def build_migration_plan(
    diff: dict[str, Any],
    roadmap_entries: list[dict[str, str]],
    implemented_entries: list[dict[str, str]] | None = None,
) -> dict[str, Any]:
    """
    Build a 4-category migration plan from a diff.json dict.

    Categories:
      update_models  — modified_schemas grouped by base name, with implemented flag
      new_resources  — new_endpoints deduplicated by normalized_path (GET as anchor)
                       where annotation != swagger_artifact
      deprecated     — removed_endpoints + removed_schemas where annotation != swagger_artifact
      roadmap_gaps   — subset of new_resources matching a ROADMAP.md Candidate/Deferred entry
    """
    if implemented_entries is None:
        implemented_entries = []

    # ---- update_models (grouped by base name, cross-referenced) ----
    raw_modified = [
        item for item in diff.get("modified_schemas", [])
        if item.get("annotation") != "swagger_artifact"
    ]
    grouped = _group_modified_schemas(raw_modified)

    update_models: list[dict[str, Any]] = []
    for g in grouped:
        impl_match = _match_implemented(g["schema_name"], implemented_entries)
        action = _action_for_modified_schema({
            "schema_name": g["schema_name"],
            "details": {
                "added_fields": g["added_fields"],
                "removed_fields": g["removed_fields"],
                "changed_fields": g["changed_fields"],
            },
        })
        update_models.append({
            "schema_name": g["schema_name"],
            "variants": g["variants"],
            "added_fields": g["added_fields"],
            "removed_fields": g["removed_fields"],
            "changed_fields": g["changed_fields"],
            "annotation": g["annotation"],
            "implemented": impl_match is not None,
            "terraform_resource": impl_match["resource_name"] if impl_match else None,
            "action": action,
        })

    # ---- new_resources (deduplicated by normalized_path) ----
    # Prefer GET as anchor; collect all methods per path
    path_methods: dict[str, list[str]] = {}
    path_meta: dict[str, dict[str, Any]] = {}
    for item in diff.get("new_endpoints", []):
        if item.get("annotation") == "swagger_artifact":
            continue
        norm = item["normalized_path"]
        method = item.get("method", "")
        path_methods.setdefault(norm, []).append(method)
        # Use GET operation as anchor; otherwise first seen
        if norm not in path_meta or method == "get":
            path_meta[norm] = item

    new_resources: list[dict[str, Any]] = []
    roadmap_gaps: list[dict[str, Any]] = []

    for norm, meta in sorted(path_meta.items()):
        methods = sorted(set(path_methods[norm]))
        roadmap_match = _match_roadmap(norm, roadmap_entries)

        entry: dict[str, Any] = {
            "normalized_path": norm,
            "methods": methods,
            "annotation": meta.get("annotation", "needs_verification"),
            "roadmap_status": roadmap_match["status"] if roadmap_match else None,
            "action": _action_for_new_resource(norm),
        }
        new_resources.append(entry)

        if roadmap_match:
            roadmap_gaps.append({
                "normalized_path": norm,
                "roadmap_entry": roadmap_match["api_section"],
                "roadmap_status": roadmap_match["status"],
                "annotation": meta.get("annotation", "needs_verification"),
                "action": "New endpoint matches ROADMAP Candidate — schedule implementation",
            })

    # ---- deprecated ----
    deprecated: list[dict[str, Any]] = []
    for item in diff.get("removed_endpoints", []):
        if item.get("annotation") == "swagger_artifact":
            continue
        deprecated.append({
            "type": "endpoint",
            "normalized_path": item["normalized_path"],
            "method": item.get("method", ""),
            "annotation": item.get("annotation", "needs_verification"),
            "action": f"Remove or update client code for {item['method'].upper()} {item['normalized_path']}",
        })
    for item in diff.get("removed_schemas", []):
        if item.get("annotation") == "swagger_artifact":
            continue
        deprecated.append({
            "type": "schema",
            "schema_name": item["schema_name"],
            "annotation": item.get("annotation", "needs_verification"),
            "action": f"Remove {item['schema_name']} struct and all usages",
        })

    impl_count = sum(1 for m in update_models if m["implemented"])

    plan: dict[str, Any] = {
        "generated_from": {
            "old_version": diff.get("old_version", "unknown"),
            "new_version": diff.get("new_version", "unknown"),
        },
        "summary": {
            "update_models": len(update_models),
            "update_models_implemented": impl_count,
            "new_resources": len(new_resources),
            "deprecated": len(deprecated),
            "roadmap_gaps": len(roadmap_gaps),
        },
        "update_models": update_models,
        "new_resources": new_resources,
        "deprecated": deprecated,
        "roadmap_gaps": roadmap_gaps,
    }
    return plan


# ---------------------------------------------------------------------------
# Markdown renderer
# ---------------------------------------------------------------------------

def _md_table(headers: list[str], rows: list[list[str]]) -> str:
    if not rows:
        return "_None_\n"
    sep = ["---"] * len(headers)
    lines = [
        "| " + " | ".join(headers) + " |",
        "| " + " | ".join(sep) + " |",
    ]
    for row in rows:
        lines.append("| " + " | ".join(str(c).replace("|", "\\|") for c in row) + " |")
    return "\n".join(lines) + "\n"


def render_markdown(plan: dict[str, Any]) -> str:
    gf = plan["generated_from"]
    s = plan["summary"]

    lines: list[str] = [
        f"# FlashBlade Migration Plan: {gf['old_version']} → {gf['new_version']}",
        "",
        "## Summary",
        "",
        "| Category | Count |",
        "| -------- | ----- |",
        f"| Model updates | {s['update_models']} ({s['update_models_implemented']} impact implemented resources) |",
        f"| New resources | {s['new_resources']} |",
        f"| Deprecated | {s['deprecated']} |",
        f"| Roadmap gaps | {s['roadmap_gaps']} |",
        "",
    ]

    # update_models — sorted: implemented first, then not implemented
    sorted_models = sorted(plan["update_models"], key=lambda m: (not m.get("implemented", False), m["schema_name"]))
    lines += ["## Model Updates", ""]
    rows = [
        [
            item["schema_name"],
            ", ".join(item["added_fields"]) or "—",
            ", ".join(item["removed_fields"]) or "—",
            "Yes" if item.get("implemented") else "No",
            item.get("terraform_resource") or "—",
            item["action"],
        ]
        for item in sorted_models
    ]
    lines.append(_md_table(["Schema", "Added Fields", "Removed Fields", "Implemented", "Resource", "Action"], rows))

    # new_resources
    lines += ["## New Resources", ""]
    rows = [
        [
            f"`{item['normalized_path']}`",
            ", ".join(item["methods"]),
            item["annotation"],
            item["roadmap_status"] or "—",
            item["action"],
        ]
        for item in plan["new_resources"]
    ]
    lines.append(_md_table(["Path", "Methods", "Annotation", "Roadmap", "Action"], rows))

    # deprecated
    lines += ["## Deprecated", ""]
    rows = [
        [
            item.get("normalized_path") or item.get("schema_name", "—"),
            item.get("method", item.get("type", "—")),
            item["annotation"],
            item["action"],
        ]
        for item in plan["deprecated"]
    ]
    lines.append(_md_table(["Path/Schema", "Method/Type", "Annotation", "Action"], rows))

    # roadmap_gaps
    lines += ["## Roadmap Gaps", ""]
    rows = [
        [
            f"`{item['normalized_path']}`",
            item["roadmap_entry"],
            item["roadmap_status"],
            item["annotation"],
            item["action"],
        ]
        for item in plan["roadmap_gaps"]
    ]
    lines.append(_md_table(["Path", "ROADMAP Entry", "Status", "Annotation", "Action"], rows))

    # action items (non-swagger_artifact)
    all_items = (
        plan["update_models"]
        + plan["new_resources"]
        + plan["deprecated"]
        + plan["roadmap_gaps"]
    )
    action_rows = [
        [item.get("normalized_path") or item.get("schema_name", "—"), item["action"]]
        for item in all_items
        if item.get("annotation") != "swagger_artifact"
    ]

    lines += ["## Action Items", ""]
    lines.append(_md_table(["Item", "Action"], action_rows))

    return "\n".join(lines)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> int:
    parser = argparse.ArgumentParser(
        description="Turn a diff.json (from diff_swagger.py) into an actionable migration plan."
    )
    parser.add_argument("diff_json", metavar="diff.json", help="Path to diff JSON produced by diff_swagger.py")
    parser.add_argument("roadmap_md", metavar="ROADMAP.md", help="Path to ROADMAP.md for cross-reference")
    parser.add_argument("--output", help="Write result to file (default: stdout)")
    parser.add_argument(
        "--format",
        choices=["json", "markdown"],
        default="json",
        help="Output format (default: json)",
    )
    args = parser.parse_args()

    # Validate inputs
    if not Path(args.diff_json).is_file():
        print(f"error: file not found: {args.diff_json}", file=sys.stderr)
        return 1
    if not Path(args.roadmap_md).is_file():
        print(f"error: file not found: {args.roadmap_md}", file=sys.stderr)
        return 1

    with open(args.diff_json, encoding="utf-8") as fh:
        diff = json.load(fh)

    roadmap_entries = _parse_roadmap_not_implemented(args.roadmap_md)
    implemented_entries = _parse_roadmap_implemented(args.roadmap_md)

    plan = build_migration_plan(diff, roadmap_entries, implemented_entries)

    if args.format == "markdown":
        output = render_markdown(plan)
    else:
        output = json.dumps(plan, indent=2)

    if args.output:
        Path(args.output).write_text(output, encoding="utf-8")
        print(f"Written to {args.output}", file=sys.stderr)
    else:
        print(output)

    return 0


if __name__ == "__main__":
    sys.exit(main())
