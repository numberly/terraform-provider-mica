"""
Test suite for swagger_utils.py.

Run from project root:
  python3 -m pytest .claude/skills/_shared/test_swagger_utils.py -v
"""
from __future__ import annotations
import json
import pathlib
import pytest

from _shared.swagger_utils import resolve_all_of, normalize_path, flatten_schema

# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

_PROJECT_ROOT = pathlib.Path(__file__).parents[3]  # .claude/skills/_shared/ → project root
_SWAGGER_PATH = _PROJECT_ROOT / "swagger-2.22.json"


@pytest.fixture(scope="module")
def swagger_dict():
    with open(_SWAGGER_PATH) as f:
        return json.load(f)


@pytest.fixture(scope="module")
def resolved(swagger_dict):
    return resolve_all_of(swagger_dict)


# ---------------------------------------------------------------------------
# normalize_path
# ---------------------------------------------------------------------------

def test_normalize_path_with_version():
    assert normalize_path("/api/2.22/buckets") == "buckets"


def test_normalize_path_active_directory():
    assert normalize_path("/api/2.22/active-directory") == "active-directory"


def test_normalize_path_no_version():
    assert normalize_path("/api/login") == "login"


def test_normalize_path_oauth2():
    assert normalize_path("/oauth2/1.0/token") == "token"


def test_normalize_path_no_prefix():
    assert normalize_path("buckets") == "buckets"


# ---------------------------------------------------------------------------
# resolve_all_of
# ---------------------------------------------------------------------------

def test_resolve_all_of_returns_dict(resolved):
    assert isinstance(resolved, dict)
    # 709 total, minus 101 private, minus 292 wrappers; expect >200 non-wrapper non-private schemas
    assert len(resolved) > 200


def test_resolve_all_of_no_allof_in_output(resolved):
    schemas_with_allof = [name for name, schema in resolved.items() if "allOf" in schema]
    assert schemas_with_allof == [], f"Schemas still containing allOf: {schemas_with_allof[:5]}"


def test_resolve_all_of_no_ref_in_output(resolved):
    schemas_with_ref = [name for name, schema in resolved.items() if "$ref" in schema]
    assert schemas_with_ref == [], f"Schemas still containing $ref: {schemas_with_ref[:5]}"


def test_resolve_all_of_skips_wrappers(resolved):
    assert "ActiveDirectoryGetResponse" not in resolved
    assert "BucketGetResponse" not in resolved


def test_resolve_all_of_skips_private(resolved):
    private = [k for k in resolved if k.startswith("_")]
    assert private == [], f"Private schemas in output: {private}"


def test_resolve_all_of_bucket_has_properties(resolved):
    assert "Bucket" in resolved
    props = resolved["Bucket"].get("properties", {})
    assert isinstance(props, dict) and len(props) > 0


# ---------------------------------------------------------------------------
# flatten_schema
# ---------------------------------------------------------------------------

def test_flatten_schema_allof(resolved):
    schema = {"allOf": [{"$ref": "#/components/schemas/Bucket"}]}
    result = flatten_schema(schema, resolved)
    assert isinstance(result["properties"], dict)
    assert len(result["properties"]) > 0


def test_flatten_schema_plain(resolved):
    schema = {"type": "object", "properties": {"name": {"type": "string"}}}
    result = flatten_schema(schema, resolved)
    assert result["properties"] == {"name": {"type": "string"}}
    assert result["type"] == "object"


def test_flatten_schema_required_dedup(resolved):
    schema = {
        "allOf": [
            {"type": "object", "required": ["name"], "properties": {"name": {"type": "string"}}},
            {"type": "object", "required": ["name"], "properties": {"name": {"type": "string"}}},
        ]
    }
    result = flatten_schema(schema, resolved)
    assert result["required"].count("name") == 1


def test_flatten_schema_empty(resolved):
    result = flatten_schema({}, resolved)
    assert result == {"properties": {}, "required": [], "description": "", "type": ""}
