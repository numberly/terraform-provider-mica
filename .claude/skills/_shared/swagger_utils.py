"""
swagger_utils.py — Shared utilities for FlashBlade swagger.json processing.

Functions:
  resolve_all_of(swagger_dict)  → dict[str, dict]
  normalize_path(path)          → str
  flatten_schema(schema, resolved_schemas) → dict

stdlib only: json, re, pathlib, collections, typing
"""
from __future__ import annotations
import re
from typing import Any


# ---------------------------------------------------------------------------
# resolve_all_of
# ---------------------------------------------------------------------------

_WRAPPER_SUFFIXES = ("Response", "GetResponse")


def resolve_all_of(swagger_dict: dict[str, Any]) -> dict[str, dict]:
    """
    Resolve all schemas in swagger_dict['components']['schemas'], merging allOf
    chains recursively. Returns a dict of schema_name → flattened schema.

    Skips wrapper schemas (*Response, *GetResponse).
    Private schemas (_-prefixed) are resolved inline but not emitted standalone.
    """
    raw_schemas: dict[str, dict] = swagger_dict.get("components", {}).get("schemas", {})

    def _resolve(schema: dict, depth: int = 0) -> dict:
        if depth > 20:
            raise ValueError("resolve_all_of: max recursion depth exceeded (circular ref?)")
        if "$ref" in schema:
            ref_name = schema["$ref"].split("/")[-1]
            return _resolve(raw_schemas.get(ref_name, {}), depth + 1)
        if "allOf" in schema:
            return _merge([_resolve(entry, depth + 1) for entry in schema["allOf"]])
        # Resolve allOf inside properties recursively
        result = dict(schema)
        if "properties" in schema:
            result["properties"] = {
                k: _resolve(v, depth + 1) if ("allOf" in v or "$ref" in v) else v
                for k, v in schema["properties"].items()
            }
        return result

    def _merge(schemas: list[dict]) -> dict:
        merged: dict[str, Any] = {}
        properties: dict[str, Any] = {}
        required: list[str] = []
        for s in schemas:
            if "properties" in s:
                properties.update(s["properties"])
            if "required" in s:
                for r in s["required"]:
                    if r not in required:
                        required.append(r)
            for k, v in s.items():
                if k not in ("properties", "required", "allOf", "$ref"):
                    if k not in merged:
                        merged[k] = v
        if properties:
            merged["properties"] = properties
        if required:
            merged["required"] = required
        return merged

    result: dict[str, dict] = {}
    for name, schema in raw_schemas.items():
        # Skip private schemas from standalone output
        if name.startswith("_"):
            continue
        # Skip wrapper schemas
        if any(name.endswith(suffix) for suffix in _WRAPPER_SUFFIXES):
            continue
        result[name] = _resolve(schema)

    return result


# ---------------------------------------------------------------------------
# normalize_path
# ---------------------------------------------------------------------------

# Matches /api/<version>/ or /oauth2/<version>/ where version is digits.digits
_API_PREFIX_RE = re.compile(r"^/(?:api|oauth2)/\d+\.\d+/")
# Matches /api/ with no version (e.g. /api/login)
_API_NO_VERSION_RE = re.compile(r"^/api/")
# Matches /oauth2/ with no version
_OAUTH2_NO_VERSION_RE = re.compile(r"^/oauth2/")


def normalize_path(path: str) -> str:
    """
    Strip /api/<version>/ or /oauth2/<version>/ prefix from a swagger path.

    Examples:
      /api/2.22/buckets        → buckets
      /api/login               → login
      /oauth2/1.0/token        → token
      /api/2.22/active-directory → active-directory
    """
    if _API_PREFIX_RE.match(path):
        return _API_PREFIX_RE.sub("", path)
    if _API_NO_VERSION_RE.match(path):
        return _API_NO_VERSION_RE.sub("", path)
    if _OAUTH2_NO_VERSION_RE.match(path):
        return _OAUTH2_NO_VERSION_RE.sub("", path)
    return path.lstrip("/")


# ---------------------------------------------------------------------------
# flatten_schema
# ---------------------------------------------------------------------------


def flatten_schema(schema: dict[str, Any], resolved_schemas: dict[str, dict]) -> dict[str, Any]:
    """
    Flatten a single schema dict, resolving $ref and allOf using resolved_schemas.

    Returns dict with keys: properties (dict), required (list), description (str), type (str).
    """

    def _lookup_ref(ref: str) -> dict:
        name = ref.split("/")[-1]
        return resolved_schemas.get(name, {})

    def _flatten_one(s: dict) -> dict:
        if "$ref" in s:
            return _lookup_ref(s["$ref"])
        if "allOf" in s:
            parts = [_flatten_one(entry) for entry in s["allOf"]]
            return _merge_flat(parts)
        return s

    def _merge_flat(parts: list[dict]) -> dict:
        props: dict[str, Any] = {}
        req: list[str] = []
        desc = ""
        typ = ""
        for p in parts:
            props.update(p.get("properties", {}))
            for r in p.get("required", []):
                if r not in req:
                    req.append(r)
            if not desc and p.get("description"):
                desc = p["description"]
            if not typ and p.get("type"):
                typ = p["type"]
        return {"properties": props, "required": req, "description": desc, "type": typ}

    flat = _flatten_one(schema)
    return {
        "properties": flat.get("properties", {}),
        "required": flat.get("required", []),
        "description": flat.get("description", ""),
        "type": flat.get("type", ""),
    }
