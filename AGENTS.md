# AGENTS.md — SZTU-autologin

> Guidelines for agentic coding agents working in this repository.

## Project Overview

SZTU (深圳技术大学) campus network auto-login tool. Go 1.23+, Windows-focused.
Single binary with no external dependencies.

## Commands

### Build
```bash
go build -o sztu-autologin.exe
```

### Run
```bash
./sztu-autologin.exe setup     # Interactive configuration
./sztu-autologin.exe login     # Login immediately
./sztu-autologin.exe daemon    # Background mode with auto-reconnect
./sztu-autologin.exe help      # Show help
```

### Autostart Management
```bash
./sztu-autologin.exe autostart on      # Enable auto-start on login
./sztu-autologin.exe autostart off     # Disable auto-start
./sztu-autologin.exe autostart status  # Check status
```

### Test
```bash
go test ./...
```

### Format
```bash
go fmt ./...
```

## Code Style

### Naming Conventions
- **Types/Structs**: PascalCase — `LoginEngine`, `Config`
- **Functions/Methods**: PascalCase (exported) or camelCase (unexported)
- **Variables**: camelCase — `checkInterval`, `autoReconnect`
- **Constants**: PascalCase or UPPER_SNAKE_CASE — `DefaultConfig`, `TASK_NAME`

### Error Handling
- Return errors as second return value
- Use `fmt.Errorf` with `%w` for error wrapping
- Main entry points should handle errors and exit with appropriate codes

### Comments
- Exported functions should have doc comments
- Chinese comments acceptable for domain-specific logic

### Module Structure
```
main.go              # Entry point, command routing
config.go            # Configuration management
login.go             # SRUN portal login logic
daemon.go            # Background auto-reconnect
cmd_setup.go         # Interactive setup command
utils.go             # Crypto utilities (XEncode, MD5, SHA1, Base64)
autostart_windows.go # Windows Task Scheduler integration
```

### Configuration
- Config stored in `config.json` (gitignored) next to executable
- Password stored in plaintext (user decision)
- Use `LoadConfig()` and `SaveConfig()` for config operations

### Platform
- Windows-only (uses `schtasks` for auto-start)
- Shell: PowerShell 7
- Build tag `//go:build windows` for Windows-specific code
