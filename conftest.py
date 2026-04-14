"""
conftest.py — Root conftest: add .claude/skills/ to sys.path so _shared is importable.
"""
import sys
import pathlib

_skills_dir = str(pathlib.Path(__file__).parent / ".claude" / "skills")
if _skills_dir not in sys.path:
    sys.path.insert(0, _skills_dir)
