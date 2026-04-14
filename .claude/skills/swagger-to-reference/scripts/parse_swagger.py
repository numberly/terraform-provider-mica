"""
parse_swagger.py — Convert FlashBlade swagger.json to AI-optimized markdown reference.

Usage:
    python3 parse_swagger.py swagger-2.22.json [--output PATH] [--version VERSION]

stdlib only: json, sys, pathlib, argparse, re, collections
"""
from __future__ import annotations

import sys
import pathlib

# MANDATORY: path setup for shared library
sys.path.insert(0, str(pathlib.Path(__file__).parent.parent.parent))
from _shared.swagger_utils import resolve_all_of, normalize_path, flatten_schema

import argparse
import json
import re
from collections import OrderedDict
from typing import Any


# HTTP methods in canonical sort order
HTTP_METHODS_ORDER = ["get", "post", "put", "patch", "delete", "head", "options"]
HTTP_METHODS_SET = set(HTTP_METHODS_ORDER)

# Common query params that are NOT shown per-endpoint (they go in the Common Parameters table)
COMMON_PARAMS = {
    "filter", "sort", "limit", "offset", "names", "ids", "continuation_token",
    "allow_errors", "context_names", "member_ids", "member_names",
    "policy_ids", "policy_names", "X-Request-ID",
}

# Truncate descriptions at this many characters (matching FLASHBLADE_API.md behavior)
DESC_TRUNCATE = 50


class SwaggerConverter:
    def __init__(self, swagger_path: str, version: str | None = None):
        swagger_path_obj = pathlib.Path(swagger_path)
        if not swagger_path_obj.exists():
            print(f"Error: swagger file not found: {swagger_path}", file=sys.stderr)
            sys.exit(1)

        with open(swagger_path_obj, "r", encoding="utf-8") as f:
            try:
                self.swagger_dict: dict[str, Any] = json.load(f)
            except json.JSONDecodeError as e:
                print(f"Error: invalid JSON in {swagger_path}: {e}", file=sys.stderr)
                sys.exit(1)

        # Version: from arg or from swagger info
        if version:
            self.version = version
        else:
            self.version = self.swagger_dict.get("info", {}).get("version", "unknown")

        # Resolve all schemas (strips Response/GetResponse wrappers)
        self.resolved_schemas: dict[str, dict] = resolve_all_of(self.swagger_dict)

        # All swagger paths
        self.paths: dict[str, dict] = self.swagger_dict.get("paths", {})

    def convert(self) -> str:
        sections = [
            self._build_title(),
            "",
            self._build_summary(),
            "",
            self._build_auth(),
            "",
            self._build_common_params(),
            "",
            "---",
            "",
            self._build_endpoints(),
            "",
            "---",
            "",
            self._build_data_models(),
        ]
        return "\n".join(sections)

    def _build_title(self) -> str:
        return f"# FlashBlade REST API {self.version} — AI-Optimized Reference"

    def _build_summary(self) -> str:
        n_paths = len(self.paths)
        n_ops = 0
        for path_item in self.paths.values():
            for method in HTTP_METHODS_SET:
                if method in path_item:
                    n_ops += 1
        return (
            f"Base: `https://{{array}}/` | Version: {self.version} | "
            f"{n_paths} paths | {n_ops} ops | Auth: `x-auth-token` or `api-token` header"
        )

    def _build_auth(self) -> str:
        return (
            "## Auth\n"
            "\n"
            "**OAuth2 token exchange:**\n"
            "```\n"
            "POST /oauth2/1.0/token\n"
            "Content-Type: application/x-www-form-urlencoded\n"
            "grant_type=urn:ietf:params:oauth:grant-type:token-exchange"
            "&subject_token=<API_TOKEN>"
            "&subject_token_type=urn:ietf:params:oauth:token-type:jwt\n"
            "→ {access_token} → Authorization: Bearer <access_token>\n"
            "```\n"
            "\n"
            "**Session login:** `POST /api/login` (header `api-token: <TOKEN>`) → returns `x-auth-token` header\n"
            "**Session logout:** `POST /api/logout` (header `x-auth-token: <TOKEN>`)\n"
            "**API version:** `GET /api/api_version` → `{versions: [string]}`"
        )

    def _build_common_params(self) -> str:
        # Collect all query parameters from all path operations, deduplicate by name
        params_by_name: dict[str, dict] = {}
        for path_item in self.paths.values():
            # Path-level params
            for param in path_item.get("parameters", []):
                p = self._resolve_param(param)
                if p and p.get("in") == "query":
                    name = p.get("name", "")
                    if name not in params_by_name:
                        params_by_name[name] = p
            # Operation-level params
            for method in HTTP_METHODS_SET:
                if method not in path_item:
                    continue
                op = path_item[method]
                for param in op.get("parameters", []):
                    p = self._resolve_param(param)
                    if p and p.get("in") == "query":
                        name = p.get("name", "")
                        if name not in params_by_name:
                            params_by_name[name] = p

        lines = [
            "## Common Parameters",
            "",
            "Most list (GET) endpoints support these query params:",
            "",
            "| Param | Type | Description |",
            "|-------|------|-------------|",
        ]

        for name in sorted(params_by_name.keys()):
            p = params_by_name[name]
            schema = p.get("schema", {})
            ptype = self._format_type(schema)
            desc = p.get("description", "")
            if schema.get("default") is not None:
                desc = f"{ptype} (default: {schema['default']}) {desc}".strip()
                # rebuild with type embedded
                ptype = "boolean" if schema.get("type") == "boolean" else ptype
                desc = p.get("description", "")
                # Re-format description with default
                default_val = schema["default"]
                desc_full = desc
                desc_str = f"{ptype} (default: {default_val}) | {_truncate(desc_full, 70)}"
                ptype_str = ptype
                lines.append(f"| `{name}` | {ptype_str} (default: {default_val}) | {_truncate(desc_full, 70)}... |")
            else:
                desc_trunc = _truncate(desc, 70)
                if len(desc) > 70:
                    desc_trunc += "..."
                lines.append(f"| `{name}` | {ptype} | {desc_trunc} |")

        lines.extend([
            "",
            "**Filter syntax:** `field op value`. Ops: `=`, `!=`, `<`, `>`, `<=`, `>=`, `and`, `or`, `not`.",
            "Example: `filter=name='fs1' and provisioned>1048576`",
        ])
        return "\n".join(lines)

    def _build_endpoints(self) -> str:
        # Group paths by tag
        tag_to_paths: dict[str, list[tuple[str, dict]]] = {}
        for path, path_item in self.paths.items():
            # Find first tag from first operation
            tag = "Other"
            for method in HTTP_METHODS_ORDER:
                if method in path_item:
                    op = path_item[method]
                    tags = op.get("tags", [])
                    if tags:
                        tag = tags[0]
                    break
            if tag not in tag_to_paths:
                tag_to_paths[tag] = []
            tag_to_paths[tag].append((path, path_item))

        lines = ["## Endpoints"]

        for tag in sorted(tag_to_paths.keys()):
            tag_title = _tag_to_title(tag)
            lines.append("")
            lines.append(f"### {tag_title}")
            lines.append("")

            for path, path_item in tag_to_paths[tag]:
                full_path = f"/api/{self.version}/{normalize_path(path)}"
                for method in HTTP_METHODS_ORDER:
                    if method not in path_item:
                        continue
                    op = path_item[method]
                    line = f"- **{method.upper()}** `{full_path}`"

                    # Collect non-common query params
                    non_common = []
                    all_params = list(path_item.get("parameters", []))
                    all_params.extend(op.get("parameters", []))
                    seen_params = set()
                    for param in all_params:
                        p = self._resolve_param(param)
                        if not p:
                            continue
                        if p.get("in") != "query":
                            continue
                        name = p.get("name", "")
                        if name in COMMON_PARAMS:
                            continue
                        if name in seen_params:
                            continue
                        seen_params.add(name)
                        schema = p.get("schema", {})
                        ptype = self._format_type(schema)
                        non_common.append(f"`{name}`({ptype})")

                    if non_common:
                        line += f" | Params: {', '.join(non_common)}"

                    # Request body fields
                    body_fields = self._extract_body_fields(op)
                    if body_fields:
                        fields_str = ", ".join(body_fields)
                        line += f" | Body: {fields_str}"

                    lines.append(line)

        return "\n".join(lines)

    def _build_data_models(self) -> str:
        lines = ["## Data Models (Key Resources)"]
        lines.append("")

        for schema_name in sorted(self.resolved_schemas.keys()):
            schema = self.resolved_schemas[schema_name]
            flat = flatten_schema(schema, self.resolved_schemas)
            props = flat.get("properties", {})
            if not props:
                continue

            field_parts = []
            for field_name in sorted(props.keys()):
                prop = props[field_name]
                is_ro = prop.get("readOnly", False)
                ptype = self._format_type(prop)
                ro_prefix = "ro " if is_ro else ""
                desc = prop.get("description", "")
                desc_trunc = desc[:DESC_TRUNCATE] if len(desc) > DESC_TRUNCATE else desc
                field_parts.append(f"`{field_name}`({ro_prefix}{ptype}): {desc_trunc}")

            if field_parts:
                line = f"**{schema_name}**: {' | '.join(field_parts)}"
                lines.append(line)

        return "\n".join(lines)

    def _resolve_param(self, param: dict) -> dict | None:
        """Resolve a parameter (handles $ref)."""
        if "$ref" in param:
            ref_name = param["$ref"].split("/")[-1]
            components = self.swagger_dict.get("components", {})
            parameters = components.get("parameters", {})
            return parameters.get(ref_name)
        return param

    def _extract_body_fields(self, op: dict) -> list[str]:
        """Extract request body fields as formatted strings."""
        request_body = op.get("requestBody", {})
        if not request_body:
            return []

        content = request_body.get("content", {})
        json_content = content.get("application/json", {})
        schema = json_content.get("schema", {})

        if not schema:
            return []

        flat = flatten_schema(schema, self.resolved_schemas)
        props = flat.get("properties", {})

        fields = []
        for field_name in sorted(props.keys()):
            prop = props[field_name]
            is_ro = prop.get("readOnly", False)
            ptype = self._format_type(prop)
            ro_prefix = "ro " if is_ro else ""
            fields.append(f"`{field_name}`({ro_prefix}{ptype})")

        return fields

    def _format_type(self, prop_schema: dict) -> str:
        """Return OpenAPI type string from a property schema."""
        if not prop_schema:
            return "object"
        if "$ref" in prop_schema:
            return "object"
        t = prop_schema.get("type")
        if t:
            return t
        if "allOf" in prop_schema or "properties" in prop_schema:
            return "object"
        return "object"


def _truncate(s: str, n: int) -> str:
    """Truncate string to at most n characters (no ellipsis added here)."""
    return s[:n]


def _tag_to_title(tag: str) -> str:
    """Convert kebab-case tag to title-case display string."""
    return " ".join(word.capitalize() for word in tag.split("-"))


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Convert FlashBlade swagger.json to AI-optimized markdown reference."
    )
    parser.add_argument("swagger_file", help="Path to swagger.json")
    parser.add_argument(
        "--output",
        help="Output file path (default: api_references/{version}.md relative to swagger_file parent)",
        default=None,
    )
    parser.add_argument(
        "--version",
        help="API version string (e.g. 2.22); if omitted, extracted from swagger info.version",
        default=None,
    )
    args = parser.parse_args()

    converter = SwaggerConverter(args.swagger_file, version=args.version)

    output_path: pathlib.Path
    if args.output:
        output_path = pathlib.Path(args.output)
    else:
        swagger_parent = pathlib.Path(args.swagger_file).parent
        output_path = swagger_parent / "api_references" / f"{converter.version}.md"

    output_path.parent.mkdir(parents=True, exist_ok=True)

    content = converter.convert()
    output_path.write_text(content, encoding="utf-8")
    print(f"Written: {output_path}", file=sys.stderr)


if __name__ == "__main__":
    main()
