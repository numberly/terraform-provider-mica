#!/usr/bin/env python3
"""
upgrade_version.py — Mechanical API version-string replacement for the FlashBlade provider.

Replaces version strings in three target locations:
  1. internal/client/client.go       — const APIVersion = "OLD"
  2. internal/testmock/server.go     — last element of the versions slice
  3. internal/testmock/handlers/*.go — /api/OLD/ path prefixes in HandleFunc calls

Usage:
    python3 upgrade_version.py --from 2.22 --to 2.23 [--dry-run | --apply] [--project-root /path]
"""

import re
import sys
import argparse
from pathlib import Path
from collections import namedtuple

Change = namedtuple("Change", ["lineno", "old_line", "new_line"])
FileChange = namedtuple("FileChange", ["path", "new_content", "changes"])


def replace_client_version(content: str, old: str, new: str) -> tuple[str, list[Change]]:
    """Replace const APIVersion = "OLD" → const APIVersion = "NEW" in client.go."""
    pattern = re.compile(r'(const APIVersion = ")' + re.escape(old) + r'"')
    changes = []
    lines = content.splitlines(keepends=True)
    new_lines = []
    for i, line in enumerate(lines, start=1):
        new_line = pattern.sub(r'\g<1>' + new + '"', line)
        if new_line != line:
            changes.append(Change(lineno=i, old_line=line.rstrip("\n"), new_line=new_line.rstrip("\n")))
        new_lines.append(new_line)
    return "".join(new_lines), changes


def replace_server_versions(content: str, old: str, new: str) -> tuple[str, list[Change]]:
    """Replace "OLD" only on the line containing "versions": in server.go."""
    changes = []
    lines = content.splitlines(keepends=True)
    new_lines = []
    for i, line in enumerate(lines, start=1):
        if '"versions":' in line:
            # Only replace the last occurrence of "OLD" on this line to avoid
            # touching older version strings like "2.12" or "2.15".
            old_quoted = '"' + old + '"'
            new_quoted = '"' + new + '"'
            if old_quoted in line:
                # Replace only the last occurrence (the current/latest version)
                idx = line.rfind(old_quoted)
                new_line = line[:idx] + new_quoted + line[idx + len(old_quoted):]
                changes.append(Change(lineno=i, old_line=line.rstrip("\n"), new_line=new_line.rstrip("\n")))
                new_lines.append(new_line)
                continue
        new_lines.append(line)
    return "".join(new_lines), changes


def replace_handler_paths(content: str, old: str, new: str) -> tuple[str, list[Change]]:
    """Replace /api/OLD/ → /api/NEW/ in handler files."""
    old_prefix = f"/api/{old}/"
    new_prefix = f"/api/{new}/"
    changes = []
    lines = content.splitlines(keepends=True)
    new_lines = []
    for i, line in enumerate(lines, start=1):
        if old_prefix in line:
            new_line = line.replace(old_prefix, new_prefix)
            changes.append(Change(lineno=i, old_line=line.rstrip("\n"), new_line=new_line.rstrip("\n")))
            new_lines.append(new_line)
        else:
            new_lines.append(line)
    return "".join(new_lines), changes


def find_replacements(project_root: Path, old: str, new: str) -> list[FileChange]:
    """Discover all version-string replacements across the three target sets."""
    results = []

    # 1. internal/client/client.go
    client_go = project_root / "internal" / "client" / "client.go"
    if not client_go.exists():
        print(f"WARNING: {client_go} not found — skipping client.go replacement", file=sys.stderr)
    else:
        content = client_go.read_text(encoding="utf-8")
        new_content, changes = replace_client_version(content, old, new)
        if changes:
            results.append(FileChange(path=client_go, new_content=new_content, changes=changes))

    # 2. internal/testmock/server.go
    server_go = project_root / "internal" / "testmock" / "server.go"
    if not server_go.exists():
        print(f"WARNING: {server_go} not found — skipping server.go replacement", file=sys.stderr)
    else:
        content = server_go.read_text(encoding="utf-8")
        new_content, changes = replace_server_versions(content, old, new)
        if changes:
            results.append(FileChange(path=server_go, new_content=new_content, changes=changes))

    # 3. internal/testmock/handlers/*.go
    handlers_dir = project_root / "internal" / "testmock" / "handlers"
    if not handlers_dir.exists():
        print(f"WARNING: {handlers_dir} not found — skipping handler path replacements", file=sys.stderr)
    else:
        for handler_file in sorted(handlers_dir.glob("*.go")):
            content = handler_file.read_text(encoding="utf-8")
            new_content, changes = replace_handler_paths(content, old, new)
            if changes:
                results.append(FileChange(path=handler_file, new_content=new_content, changes=changes))

    return results


def apply_replacements(file_changes: list[FileChange]) -> None:
    """Write updated content back to each file."""
    for fc in file_changes:
        fc.path.write_text(fc.new_content, encoding="utf-8")


def _format_snippet(line: str, max_len: int = 80) -> str:
    stripped = line.strip()
    if len(stripped) > max_len:
        return stripped[:max_len] + "…"
    return stripped


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Replace FlashBlade API version strings across the provider codebase."
    )
    parser.add_argument("--from", dest="from_version", required=True, metavar="VERSION",
                        help="Current API version string (e.g. 2.22)")
    parser.add_argument("--to", dest="to_version", required=True, metavar="VERSION",
                        help="Target API version string (e.g. 2.23)")
    parser.add_argument("--project-root", default=None, metavar="DIR",
                        help="Root of the provider repository (default: cwd)")

    mode_group = parser.add_mutually_exclusive_group()
    mode_group.add_argument("--dry-run", dest="dry_run", action="store_true", default=True,
                            help="Print replacements without modifying files (default)")
    mode_group.add_argument("--apply", dest="apply", action="store_true", default=False,
                            help="Apply replacements to files")

    args = parser.parse_args()

    # When --apply is given, dry_run should be False.
    dry_run = not args.apply

    project_root = Path(args.project_root) if args.project_root else Path.cwd()
    old_version = args.from_version
    new_version = args.to_version

    file_changes = find_replacements(project_root, old_version, new_version)

    if not file_changes:
        print(f"No replacements found for version {old_version}", file=sys.stderr)
        sys.exit(1)

    if dry_run:
        print(f"DRY RUN — no files modified")
        print(f"Would replace {old_version!r} → {new_version!r} in {len(file_changes)} file(s):\n")
    else:
        print(f"APPLYING — replacing {old_version!r} → {new_version!r} in {len(file_changes)} file(s):\n")

    for fc in file_changes:
        for ch in fc.changes:
            old_snippet = _format_snippet(ch.old_line)
            new_snippet = _format_snippet(ch.new_line)
            print(f"{fc.path}:{ch.lineno}: {old_snippet} → {new_snippet}")

    if not dry_run:
        apply_replacements(file_changes)
        total_files = len(file_changes)
        print(f"\nDone. {total_files} file(s) modified.")


if __name__ == "__main__":
    main()
