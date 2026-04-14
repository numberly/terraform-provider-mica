"""
browse_api.py — Browse a FlashBlade API markdown reference file.

Usage:
    python3 browse_api.py <reference.md> --tag TAG
    python3 browse_api.py <reference.md> --method METHOD
    python3 browse_api.py <reference.md> --schema SCHEMANAME
    python3 browse_api.py <reference.md> --compare SCHEMA1 SCHEMA2
    python3 browse_api.py <reference.md> --stats
    python3 browse_api.py <reference.md> --search PATTERN

stdlib only: sys, pathlib, argparse, re, collections
"""
from __future__ import annotations

import sys
import pathlib

sys.path.insert(0, str(pathlib.Path(__file__).parent.parent.parent))

import argparse
import re
from collections import defaultdict


# Regex patterns for parsing
_RE_ENDPOINT = re.compile(r"^- \*\*(\w+)\*\*\s+`([^`]+)`")
_RE_SCHEMA_START = re.compile(r"^\*\*([A-Za-z0-9_]+)\*\*:\s*(.*)")
_RE_FIELD = re.compile(r"`(\w+)`\((ro )?([^)]+)\):\s*(.*)")


class ReferenceParser:
    def __init__(self, ref_path: str):
        p = pathlib.Path(ref_path)
        if not p.exists():
            print(f"Error: reference file not found: {ref_path}", file=sys.stderr)
            sys.exit(1)
        text = p.read_text(encoding="utf-8")
        self._endpoints: list[dict] = []
        self._schemas: dict[str, list[dict]] = {}
        self._parse(text)

    def _parse(self, text: str) -> None:
        # Section states: none, endpoints, models
        section = "none"
        current_tag = ""
        # For multi-line schema entries: accumulate continuation lines
        pending_schema_name: str = ""
        pending_schema_text: str = ""

        lines = text.splitlines()
        for line in lines:
            # Detect top-level section headers
            if line.startswith("## Endpoints"):
                section = "endpoints"
                current_tag = ""
                continue
            if line.startswith("## Data Models"):
                # flush any pending schema
                if pending_schema_name:
                    self._parse_schema_line(pending_schema_name, pending_schema_text)
                    pending_schema_name = ""
                    pending_schema_text = ""
                section = "models"
                continue
            if line.startswith("## ") and section in ("endpoints", "models"):
                # Another top-level section — stop processing
                if pending_schema_name:
                    self._parse_schema_line(pending_schema_name, pending_schema_text)
                    pending_schema_name = ""
                    pending_schema_text = ""
                section = "none"
                continue

            if section == "endpoints":
                # Tag subsection header
                if line.startswith("### "):
                    current_tag = line[4:].strip()
                    continue
                # Endpoint line
                m = _RE_ENDPOINT.match(line)
                if m and current_tag:
                    method = m.group(1).upper()
                    path = m.group(2)
                    rest = line[m.end():].strip()
                    params = ""
                    body = ""
                    if "| Params:" in rest:
                        params_part = rest.split("| Params:", 1)[1]
                        if "| Body:" in params_part:
                            params = params_part.split("| Body:", 1)[0].strip()
                            body = params_part.split("| Body:", 1)[1].strip()
                        else:
                            params = params_part.strip()
                    elif "| Body:" in rest:
                        body = rest.split("| Body:", 1)[1].strip()
                    self._endpoints.append({
                        "method": method,
                        "path": path,
                        "tag": current_tag,
                        "params": params,
                        "body": body,
                    })

            elif section == "models":
                # Schema line — may start with **Name**: or be a continuation
                m = _RE_SCHEMA_START.match(line)
                if m:
                    # Flush previous pending schema
                    if pending_schema_name:
                        self._parse_schema_line(pending_schema_name, pending_schema_text)
                    pending_schema_name = m.group(1)
                    pending_schema_text = m.group(2)
                elif pending_schema_name and (line.startswith(" | ") or line.startswith("| ")):
                    # Continuation of previous schema entry — strip leading whitespace/pipe
                    stripped = line.strip()
                    # Remove leading "| " so it joins cleanly with " | " separator
                    if stripped.startswith("| "):
                        stripped = stripped[2:]
                    pending_schema_text += " | " + stripped
                else:
                    # Non-schema line in models section: flush pending
                    if pending_schema_name:
                        self._parse_schema_line(pending_schema_name, pending_schema_text)
                        pending_schema_name = ""
                        pending_schema_text = ""

        # Flush final pending schema
        if pending_schema_name:
            self._parse_schema_line(pending_schema_name, pending_schema_text)

    def _parse_schema_line(self, name: str, text: str) -> None:
        """Parse pipe-separated field segments into a list of field dicts."""
        fields: list[dict] = []
        segments = text.split(" | ")
        for seg in segments:
            seg = seg.strip()
            if not seg:
                continue
            m = _RE_FIELD.match(seg)
            if m:
                field_name = m.group(1)
                read_only = bool(m.group(2))
                ftype = m.group(3).strip()
                desc = m.group(4).strip()
                fields.append({
                    "name": field_name,
                    "type": ftype,
                    "read_only": read_only,
                    "description": desc,
                })
        self._schemas[name] = fields

    @property
    def endpoints(self) -> list[dict]:
        return self._endpoints

    @property
    def schemas(self) -> dict[str, list[dict]]:
        return self._schemas


# ---------------------------------------------------------------------------
# CLI subcommand implementations
# ---------------------------------------------------------------------------

def cmd_tag(ref: ReferenceParser, tag: str) -> None:
    tag_lower = tag.lower()
    matches = [e for e in ref.endpoints if tag_lower in e["tag"].lower()]
    if not matches:
        print(f"No endpoints found for tag: {tag}")
        return
    # Use tag name from first match for display
    display_tag = matches[0]["tag"]
    print(f"Tag: {display_tag}  ({len(matches)} endpoints)")
    print()
    for e in matches:
        print(f"{e['method']:<8}{e['path']}")


def cmd_method(ref: ReferenceParser, method: str) -> None:
    method_upper = method.upper()
    matches = [e for e in ref.endpoints if e["method"] == method_upper]
    matches.sort(key=lambda e: e["path"])
    print(f"Method: {method_upper}  ({len(matches)} endpoints)")
    print()
    for e in matches:
        tag_str = f"[{e['tag']}]"
        print(f"{e['method']:<8}{e['path']:<50} {tag_str}")


def cmd_search(ref: ReferenceParser, pattern: str) -> None:
    pattern_lower = pattern.lower()
    matches = [e for e in ref.endpoints if pattern_lower in e["path"].lower()]
    matches.sort(key=lambda e: e["path"])
    print(f'Pattern: "{pattern}"  ({len(matches)} matches)')
    print()
    for e in matches:
        tag_str = f"[{e['tag']}]"
        print(f"{e['method']:<8}{e['path']:<55} {tag_str}")


def cmd_schema(ref: ReferenceParser, schema_name: str) -> None:
    schemas = ref.schemas
    if schema_name not in schemas:
        print(f"Schema not found: {schema_name}", file=sys.stderr)
        # Show closest matches (case-insensitive prefix)
        prefix = schema_name.lower()
        candidates = [n for n in schemas if n.lower().startswith(prefix)]
        if not candidates:
            # Substring match fallback
            candidates = [n for n in schemas if prefix in n.lower()]
        if candidates:
            print(f"Did you mean: {', '.join(sorted(candidates)[:5])}", file=sys.stderr)
        sys.exit(1)

    fields = schemas[schema_name]
    print(f"Schema: {schema_name}  ({len(fields)} fields)")
    print()
    header = f"{'Field':<25} {'Type':<12} {'ReadOnly':<10} Description"
    sep = f"{'-----':<25} {'----':<12} {'--------':<10} -----------"
    print(header)
    print(sep)
    for f in fields:
        desc = f["description"]
        if len(desc) > 60:
            desc = desc[:60] + "..."
        ro = "yes" if f["read_only"] else "no"
        print(f"{f['name']:<25} {f['type']:<12} {ro:<10} {desc}")


def cmd_compare(ref: ReferenceParser, schema1: str, schema2: str) -> None:
    schemas = ref.schemas
    errors = []
    if schema1 not in schemas:
        errors.append(f"Schema not found: {schema1}")
    if schema2 not in schemas:
        errors.append(f"Schema not found: {schema2}")
    if errors:
        for e in errors:
            print(e, file=sys.stderr)
        sys.exit(1)

    fields1 = {f["name"]: f for f in schemas[schema1]}
    fields2 = {f["name"]: f for f in schemas[schema2]}
    all_names = sorted(set(fields1) | set(fields2))

    print(f"Comparing: {schema1} vs {schema2}")
    print()

    col1_w = max(len(schema1), 12)
    col2_w = max(len(schema2), 12)
    header = f"{'Field':<25} {schema1:<{col1_w}} {schema2:<{col2_w}}"
    sep = f"{'-----':<25} {'-' * col1_w:<{col1_w}} {'-' * col2_w:<{col2_w}}"
    print(header)
    print(sep)
    for name in all_names:
        def _fmt(f: dict | None) -> str:
            if f is None:
                return "-"
            prefix = "ro " if f["read_only"] else ""
            return f"{prefix}{f['type']}"

        v1 = _fmt(fields1.get(name))
        v2 = _fmt(fields2.get(name))
        print(f"{name:<25} {v1:<{col1_w}} {v2:<{col2_w}}")


def cmd_stats(ref: ReferenceParser) -> None:
    endpoints = ref.endpoints
    schemas = ref.schemas

    # Unique paths (de-duplicate by path)
    unique_paths = len({e["path"] for e in endpoints})
    n_schemas = len(schemas)

    # Count by method
    method_counts: dict[str, int] = defaultdict(int)
    for e in endpoints:
        method_counts[e["method"]] += 1

    # Count by tag
    tag_counts: dict[str, int] = defaultdict(int)
    for e in endpoints:
        tag_counts[e["tag"]] += 1

    print("Reference Statistics")
    print("====================")
    print()
    print(f"Paths:    {unique_paths}")
    print(f"Schemas:  {n_schemas}")
    print("Operations by method:")
    for method in sorted(method_counts.keys()):
        print(f"  {method:<8} {method_counts[method]}")
    print()
    n_tags = len(tag_counts)
    print(f"Tags ({n_tags}):")
    for tag in sorted(tag_counts.keys()):
        print(f"  {tag} ({tag_counts[tag]} endpoints)")


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(
        prog="browse_api.py",
        description="Browse a FlashBlade API markdown reference file.",
        epilog=(
            "Examples:\n"
            "  python3 browse_api.py ref.md --tag buckets\n"
            "  python3 browse_api.py ref.md --schema BucketPost\n"
            "  python3 browse_api.py ref.md --compare BucketPost BucketPatch\n"
            "  python3 browse_api.py ref.md --stats\n"
            "  python3 browse_api.py ref.md --search replication\n"
        ),
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument("reference", help="Path to the markdown reference file (.md)")
    parser.add_argument("--tag", metavar="TAG", help="Filter endpoints by tag (case-insensitive substring)")
    parser.add_argument("--method", metavar="METHOD", help="Filter endpoints by HTTP method")
    parser.add_argument("--search", metavar="PATTERN", help="Filter endpoints by path substring")
    parser.add_argument("--schema", metavar="SCHEMANAME", help="Display all fields of a schema")
    parser.add_argument("--compare", metavar=("SCHEMA1", "SCHEMA2"), nargs=2, help="Side-by-side diff of two schemas")
    parser.add_argument("--stats", action="store_true", help="Display reference statistics")

    args = parser.parse_args()

    # Validate exactly one action flag
    actions = [
        args.tag is not None,
        args.method is not None,
        args.search is not None,
        args.schema is not None,
        args.compare is not None,
        args.stats,
    ]
    if sum(actions) == 0:
        parser.print_help(sys.stderr)
        sys.exit(1)
    if sum(actions) > 1:
        print("Error: provide exactly one action flag (--tag, --method, --search, --schema, --compare, --stats)", file=sys.stderr)
        sys.exit(1)

    ref = ReferenceParser(args.reference)

    if args.tag:
        cmd_tag(ref, args.tag)
    elif args.method:
        cmd_method(ref, args.method)
    elif args.search:
        cmd_search(ref, args.search)
    elif args.schema:
        cmd_schema(ref, args.schema)
    elif args.compare:
        cmd_compare(ref, args.compare[0], args.compare[1])
    elif args.stats:
        cmd_stats(ref)


if __name__ == "__main__":
    main()
