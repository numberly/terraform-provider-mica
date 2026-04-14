"""
conftest.py — Add .claude/skills/ to sys.path so _shared is importable.
"""
import sys
import pathlib

sys.path.insert(0, str(pathlib.Path(__file__).parent))
