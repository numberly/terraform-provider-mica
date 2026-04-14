"""
diff_swagger.py — CLI tool to compare two FlashBlade swagger.json files.

Usage:
  python3 diff_swagger.py <old.json> <new.json> [--output path] [--format json|markdown] [--discrepancies path]

Output sections:
  new_endpoints, removed_endpoints, modified_endpoints
  new_schemas, removed_schemas, modified_schemas

stdlib only: json, argparse, sys, os, pathlib
"""
from __future__ import annotations
import argparse
import json
import os
import sys
from pathlib import Path
from typing import Any

# ---------------------------------------------------------------------------
# Import shared library
# ---------------------------------------------------------------------------
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "..", "_shared"))
import swagger_utils  # noqa: E402


# ---------------------------------------------------------------------------
# Endpoint map builders
# ---------------------------------------------------------------------------

HTTP_METHODS = {"get", "post", "put", "patch", "delete", "head", "options"}


def build_endpoint_map(swagger: dict[str, Any]) -> dict[tuple[str, str], dict]:
    """Return {(normalized_path, method): operation_dict}."""
    result: dict[tuple[str, str], dict] = {}
    for raw_path, path_item in swagger.get("paths", {}).items():
        norm = swagger_utils.normalize_path(raw_path)
        for method, operation in path_item.items():
            if method.lower() not in HTTP_METHODS:
                continue
            if not isinstance(operation, dict):
                continue
            result[(norm, method.lower())] = operation
    return result


# ---------------------------------------------------------------------------
# Comparison helpers
# ---------------------------------------------------------------------------

def _param_names(operation: dict) -> list[str]:
    return sorted(p.get("name", "") for p in operation.get("parameters", []))


def _has_request_body(operation: dict) -> bool:
    return "requestBody" in operation


def _response_codes(operation: dict) -> list[str]:
    return sorted(operation.get("responses", {}).keys())


def _compare_endpoints(
    old_ops: dict[tuple[str, str], dict],
    new_ops: dict[tuple[str, str], dict],
) -> tuple[list[dict], list[dict], list[dict]]:
    """Return (new_endpoints, removed_endpoints, modified_endpoints)."""
    old_keys = set(old_ops)
    new_keys = set(new_ops)

    new_endpoints: list[dict] = []
    removed_endpoints: list[dict] = []
    modified_endpoints: list[dict] = []

    for key in sorted(new_keys - old_keys):
        norm, method = key
        new_endpoints.append({
            "normalized_path": norm,
            "method": method,
            "change_type": "new_endpoint",
            "annotation": "needs_verification",
            "details": {},
        })

    for key in sorted(old_keys - new_keys):
        norm, method = key
        removed_endpoints.append({
            "normalized_path": norm,
            "method": method,
            "change_type": "removed_endpoint",
            "annotation": "needs_verification",
            "details": {},
        })

    for key in sorted(old_keys & new_keys):
        norm, method = key
        old_op = old_ops[key]
        new_op = new_ops[key]

        changes: dict[str, Any] = {}

        old_summary = old_op.get("summary", "")
        new_summary = new_op.get("summary", "")
        if old_summary != new_summary:
            changes["summary"] = {"old": old_summary, "new": new_summary}

        old_params = _param_names(old_op)
        new_params = _param_names(new_op)
        if old_params != new_params:
            changes["parameters"] = {
                "added": sorted(set(new_params) - set(old_params)),
                "removed": sorted(set(old_params) - set(new_params)),
            }

        old_body = _has_request_body(old_op)
        new_body = _has_request_body(new_op)
        if old_body != new_body:
            changes["requestBody"] = {"old": old_body, "new": new_body}

        old_responses = _response_codes(old_op)
        new_responses = _response_codes(new_op)
        if old_responses != new_responses:
            changes["responses"] = {
                "added": sorted(set(new_responses) - set(old_responses)),
                "removed": sorted(set(old_responses) - set(new_responses)),
            }

        if changes:
            modified_endpoints.append({
                "normalized_path": norm,
                "method": method,
                "change_type": "modified_endpoint",
                "annotation": "needs_verification",
                "details": changes,
            })

    return new_endpoints, removed_endpoints, modified_endpoints


def _compare_schemas(
    old_schemas: dict[str, dict],
    new_schemas: dict[str, dict],
    old_swagger: dict[str, Any],
    new_swagger: dict[str, Any],
) -> tuple[list[dict], list[dict], list[dict]]:
    """Return (new_schemas, removed_schemas, modified_schemas)."""
    old_names = set(old_schemas)
    new_names = set(new_schemas)

    new_list: list[dict] = []
    removed_list: list[dict] = []
    modified_list: list[dict] = []

    for name in sorted(new_names - old_names):
        new_list.append({
            "schema_name": name,
            "change_type": "new_schema",
            "annotation": "needs_verification",
            "details": {},
        })

    for name in sorted(old_names - new_names):
        removed_list.append({
            "schema_name": name,
            "change_type": "removed_schema",
            "annotation": "needs_verification",
            "details": {},
        })

    for name in sorted(old_names & new_names):
        old_flat = swagger_utils.flatten_schema(old_schemas[name], old_schemas)
        new_flat = swagger_utils.flatten_schema(new_schemas[name], new_schemas)

        old_props = set(old_flat.get("properties", {}).keys())
        new_props = set(new_flat.get("properties", {}).keys())

        added_fields = sorted(new_props - old_props)
        removed_fields = sorted(old_props - new_props)

        # Changed fields: same name, different type or description
        changed_fields: list[dict] = []
        for field in sorted(old_props & new_props):
            old_field = old_flat.get("properties", {}).get(field, {})
            new_field = new_flat.get("properties", {}).get(field, {})
            old_type = old_field.get("type", old_field.get("$ref", ""))
            new_type = new_field.get("type", new_field.get("$ref", ""))
            if old_type != new_type:
                changed_fields.append({
                    "field": field,
                    "old_type": old_type,
                    "new_type": new_type,
                })

        if added_fields or removed_fields or changed_fields:
            modified_list.append({
                "schema_name": name,
                "change_type": "modified_schema",
                "annotation": "needs_verification",
                "details": {
                    "added_fields": added_fields,
                    "removed_fields": removed_fields,
                    "changed_fields": changed_fields,
                },
            })

    return new_list, removed_list, modified_list


# ---------------------------------------------------------------------------
# Discrepancy overrides
# ---------------------------------------------------------------------------

def _apply_overrides(
    sections: dict[str, list[dict]],
    discrepancies_path: str | None,
) -> None:
    if not discrepancies_path:
        return
    with open(discrepancies_path, encoding="utf-8") as fh:
        disc = json.load(fh)

    overrides: list[dict] = disc.get("overrides", [])
    override_index: dict[tuple[str, str], str] = {}
    for ov in overrides:
        key = (ov.get("normalized_path", ""), ov.get("method", ""))
        override_index[key] = ov.get("annotation", "needs_verification")

    for items in sections.values():
        for item in items:
            norm = item.get("normalized_path", "")
            method = item.get("method", "")
            key = (norm, method)
            if key in override_index:
                item["annotation"] = override_index[key]


# ---------------------------------------------------------------------------
# Markdown output
# ---------------------------------------------------------------------------

def _md_endpoint_table(items: list[dict]) -> str:
    if not items:
        return "_None_\n"
    lines = ["| Path | Method | Change | Annotation | Details |",
             "| ---- | ------ | ------ | ---------- | ------- |"]
    for item in items:
        details = json.dumps(item.get("details", {}), separators=(",", ":"))
        # Escape pipe chars in JSON so table doesn't break
        details = details.replace("|", "\\|")
        lines.append(
            f"| `{item.get('normalized_path', '')}` "
            f"| {item.get('method', '').upper()} "
            f"| {item.get('change_type', '')} "
            f"| {item.get('annotation', '')} "
            f"| {details} |"
        )
    return "\n".join(lines) + "\n"


def _md_schema_table(items: list[dict]) -> str:
    if not items:
        return "_None_\n"
    lines = ["| Schema | Change | Annotation | Details |",
             "| ------ | ------ | ---------- | ------- |"]
    for item in items:
        details = json.dumps(item.get("details", {}), separators=(",", ":"))
        details = details.replace("|", "\\|")
        lines.append(
            f"| `{item.get('schema_name', '')}` "
            f"| {item.get('change_type', '')} "
            f"| {item.get('annotation', '')} "
            f"| {details} |"
        )
    return "\n".join(lines) + "\n"


def _render_markdown(diff: dict[str, Any]) -> str:
    summary = diff["summary"]
    lines = [
        f"# FlashBlade API Diff: {diff['old_version']} → {diff['new_version']}",
        "",
        "## Summary",
        "",
        "| Category | Count |",
        "| -------- | ----- |",
        f"| New endpoints | {summary['new_endpoints']} |",
        f"| Removed endpoints | {summary['removed_endpoints']} |",
        f"| Modified endpoints | {summary['modified_endpoints']} |",
        f"| New schemas | {summary['new_schemas']} |",
        f"| Removed schemas | {summary['removed_schemas']} |",
        f"| Modified schemas | {summary['modified_schemas']} |",
        "",
        "## New Endpoints",
        "",
        _md_endpoint_table(diff["new_endpoints"]),
        "## Removed Endpoints",
        "",
        _md_endpoint_table(diff["removed_endpoints"]),
        "## Modified Endpoints",
        "",
        _md_endpoint_table(diff["modified_endpoints"]),
        "## New Schemas",
        "",
        _md_schema_table(diff["new_schemas"]),
        "## Removed Schemas",
        "",
        _md_schema_table(diff["removed_schemas"]),
        "## Modified Schemas",
        "",
        _md_schema_table(diff["modified_schemas"]),
    ]
    return "\n".join(lines)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> int:
    parser = argparse.ArgumentParser(
        description="Compare two FlashBlade swagger.json files and emit a structured diff."
    )
    parser.add_argument("old_swagger", help="Path to old swagger.json")
    parser.add_argument("new_swagger", help="Path to new swagger.json")
    parser.add_argument("--output", help="Write result to file (default: stdout)")
    parser.add_argument(
        "--format",
        choices=["json", "markdown"],
        default="json",
        help="Output format (default: json)",
    )
    parser.add_argument(
        "--discrepancies",
        help="Path to JSON file with annotation overrides",
    )
    args = parser.parse_args()

    # Load files
    for path in (args.old_swagger, args.new_swagger):
        if not Path(path).is_file():
            print(f"error: file not found: {path}", file=sys.stderr)
            return 1

    with open(args.old_swagger, encoding="utf-8") as fh:
        old_swagger = json.load(fh)
    with open(args.new_swagger, encoding="utf-8") as fh:
        new_swagger = json.load(fh)

    # Build maps
    old_endpoints = build_endpoint_map(old_swagger)
    new_endpoints = build_endpoint_map(new_swagger)

    old_schemas = swagger_utils.resolve_all_of(old_swagger)
    new_schemas = swagger_utils.resolve_all_of(new_swagger)

    # Compute diffs
    new_ep, removed_ep, modified_ep = _compare_endpoints(old_endpoints, new_endpoints)
    new_sc, removed_sc, modified_sc = _compare_schemas(
        old_schemas, new_schemas, old_swagger, new_swagger
    )

    diff: dict[str, Any] = {
        "old_version": old_swagger.get("info", {}).get("version", "unknown"),
        "new_version": new_swagger.get("info", {}).get("version", "unknown"),
        "summary": {
            "new_endpoints": len(new_ep),
            "removed_endpoints": len(removed_ep),
            "modified_endpoints": len(modified_ep),
            "new_schemas": len(new_sc),
            "removed_schemas": len(removed_sc),
            "modified_schemas": len(modified_sc),
        },
        "new_endpoints": new_ep,
        "removed_endpoints": removed_ep,
        "modified_endpoints": modified_ep,
        "new_schemas": new_sc,
        "removed_schemas": removed_sc,
        "modified_schemas": modified_sc,
    }

    # Apply discrepancy overrides
    sections = {k: diff[k] for k in (
        "new_endpoints", "removed_endpoints", "modified_endpoints",
        "new_schemas", "removed_schemas", "modified_schemas",
    )}
    _apply_overrides(sections, args.discrepancies)

    # Render
    if args.format == "markdown":
        output = _render_markdown(diff)
    else:
        output = json.dumps(diff, indent=2)

    if args.output:
        Path(args.output).write_text(output, encoding="utf-8")
        print(f"Written to {args.output}", file=sys.stderr)
    else:
        print(output)

    return 0


if __name__ == "__main__":
    sys.exit(main())
