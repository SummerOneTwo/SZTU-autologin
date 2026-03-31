# AGENTS.md — SZTU-autologin

> Guidelines for agentic coding agents working in this repository.

## Project Overview

SZTU (深圳技术大学) campus network auto-login tool. Python 3.10+, Windows-focused.
Uses `uv` for dependency management. No test framework currently exists.

## Commands

### Run
```bash
uv run main.py            # Interactive mode
uv run main.py --silent   # Silent/background mode
```

### Lint
```bash
uv run ruff check .       # Lint all files
uv run ruff check --fix . # Auto-fix lint issues
```

### Format
```bash
uv run ruff format .      # Format all files
```

### Build (PyInstaller)
```bash
uv run build.py                              # Install deps + build exe
uv run pyinstaller -F -w --name SZTU-Autologin main.py  # Direct build
```

### Dependencies
```bash
uv sync           # Install/sync dependencies
uv add <package>  # Add a dependency
```

### Running a Single Test
No test framework exists yet. To add tests, create files matching `test_*.py` or `*_test.py` and use `pytest` (add via `uv add --dev pytest`). Run with:
```bash
uv run pytest tests/test_file.py -v
```

## Code Style

### Imports
Order: stdlib → third-party → local modules. Use blank line between groups.
```python
import json
import socket
from pathlib import Path

import requests
from colorama import Fore

from core.config import ConfigManager
from utils import network
```

### Naming Conventions
- **Classes**: PascalCase — `LoginEngine`, `ConfigManager`
- **Functions/Methods**: snake_case — `get_token()`, `is_logged_in()`
- **Variables**: snake_case — `check_interval`, `auto_reconnect`
- **Constants**: UPPER_SNAKE_CASE — `DEFAULT_CONFIG`, `TASK_NAME`
- **Private helpers**: leading underscore — `_get_chksum()`, `_validate_and_fix()`

### Type Hints
Use type hints on all function signatures and class attributes:
```python
def login(self) -> LoginResult:
def is_logged_in(self) -> bool:
def __init__(self, config_file: str = "config.json"):
```

### Error Handling
- Use `try/except` blocks; return `False`/`""`/`None` on failure for utility functions
- Login operations return `LoginResult(success, message)` instead of raising
- Catch broad `Exception` in outer loops; log and continue
- Use `raise` for unrecoverable errors (e.g., `IOError` on config save failure)

### Logging
Use the `core.logger.Logger` class for all logging:
```python
self.logger.info("message")
self.logger.warning("message")
self.logger.error("message", exc)  # pass exception for stack trace
self.logger.debug("message")
```

### Docstrings
Chinese docstrings for public methods:
```python
def get_local_ip() -> str:
    """获取本机IP地址"""
```

### Line Length & Formatting
- Ruff handles formatting (`uv run ruff format .`)
- Ruff lint rules: all default except `E741` (single-letter var names allowed for crypto code)
- No explicit line length limit configured; keep lines reasonable

### Class Design
- Use `@classmethod` for factory/utility methods (see `TaskScheduler`)
- Implement `__enter__`/`__exit__` for context managers (see `FileLock`)
- Keep modules focused: `core/` for business logic, `utils/` for helpers

### Module Structure
```
core/       # Business logic (login, config, scheduling, etc.)
utils/      # Low-level helpers (crypto, network)
main.py     # CLI entry point
build.py    # PyInstaller build script
autologin.py # Legacy script (kept for reference, not imported)
```

### Configuration
- Config stored in `config.json` (gitignored)
- Use `ConfigManager` for load/save/validate — never read/write config directly
- `DEFAULT_CONFIG` defines the schema; `_validate_and_fix()` auto-repairs missing keys

### Platform
- Windows-only (uses `schtasks`, `ctypes.windll.kernel32`)
- Shell: PowerShell 7
- Lock file prevents concurrent instances
