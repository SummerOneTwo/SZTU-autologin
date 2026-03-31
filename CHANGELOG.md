# Changelog

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
