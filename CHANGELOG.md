# Changelog

## [3.2.0] - 2026-04-10

### Features
- **自动重连与守护进程状态同步**：配置项 `auto_reconnect` 现在直接反映守护进程运行状态
- 菜单开启自动重连时自动启动后台静默守护进程
- 菜单关闭自动重连时自动停止守护进程
- 交互菜单启动时检查配置与进程状态一致性，自动修正异常状态
- 开机自启动时先将配置改为 `true`，再启动守护进程

### Design Rationale
- 解决用户困惑：之前"开启自动重连"只修改配置，不启动守护进程
- 使用 `cmd /c start /b` 启动独立子进程，确保父进程退出后守护进程继续运行
- 守护进程正常退出时通过 defer 自动将配置改为 `false`
- 异常退出（如被 taskkill 强制终止）通过交互菜单启动时的一致性检查来修正

### Notes & Caveats
- `taskkill /f` 强制终止不会触发 defer，配置需通过交互菜单修正
- 新增 `autostart-launch` 内部命令供开机自启动调用

## [3.1.0] - 2026-04-07

### Features
- 后台静默运行：守护进程支持 `-hide` 参数隐藏控制台窗口
- 新增 `stop` 命令：通过进程名查找并终止守护进程
- 开机自启任务自动添加 `-hide` 参数，实现完全无感知后台运行

### Design Rationale
- 使用 Windows API `ShowWindow(GetConsoleWindow(), SW_HIDE)` 隐藏窗口
- `-hide` 参数仅在 `daemon` 命令时生效，避免交互模式误用
- `stopDaemon()` 使用 `tasklist`/`taskkill` 实现，跨平台兼容通过编译标签处理

### Notes & Caveats
- 用户手动运行 `daemon` 时窗口可见，便于调试
- 开机自启后守护进程完全静默，无窗口弹出

## [3.0.1] - 2026-04-04

### Features
- 新增交互式主菜单状态收口，统一展示网络、开机自启与自动重连状态
- 开机自启权限提升兼容中英文拒绝访问提示

### Design Rationale
- 将菜单状态采样集中到单次读取，减少重复网络探测和计划任务查询，避免交互层继续膨胀
- 发布前清理测试文件，避免仓库继续携带个人信息
- 文档同步到当前实际行为，减少 README、AGENTS、CLAUDE 之间的描述漂移

### Notes & Caveats
- `go test ./...` 在当前机器上曾被系统 `GOCACHE` 权限阻断；可使用项目内 `.gocache` 作为临时缓存目录

## [3.0.0] - 2026-04-01

### Features
- **Go 重构** - 从 Python 完全重写为 Go
- 单文件二进制，无外部依赖，约 9MB
- 交互式配置命令 (`setup`)
- 立即登录命令 (`login`)
- 后台守护进程，自动检测重连 (`daemon`)
- Windows 开机自启动管理 (`autostart`)
- 密码明文存储（用户决定，可手动编辑 config.json）

### Design Rationale
- Go 编译为单文件，部署简单，启动快
- 使用 Windows 任务计划程序实现开机自启
- SRUN 门户登录使用 XXTEA (XEncode) 加密算法
- 移除 Python 依赖和 PyInstaller 打包臃肿问题

### Notes & Caveats
- 配置文件格式与 Python 版本不兼容，需重新运行 `setup`
- 不再支持加密存储密码（简化实现）
- 不再包含日志系统（守护进程直接输出到控制台）

---

## [2.0.0] - 2026-03-31

### Features
- 重构为模块化架构
- 新增配置文件管理账号密码
- 新增宿舍区/教学区可切换
- 新增命令行交互式菜单
- 新增定时检测网络，掉网自动重连
- 新增Windows任务计划开机自启
- 新增后台静默运行模式
- 新增完整的日志系统
- 新增单实例锁防止重复运行

### Design Rationale
- 使用JSON配置简洁易读
- HTTP方式检测校园网连接（支持有线/无线）
- Windows任务计划确保开机自启稳定可靠
- 模块化设计便于维护和扩展
