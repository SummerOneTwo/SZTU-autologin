# SZTU 校园网自动登录工具 v2.0 设计文档

**日期**: 2026-03-31
**版本**: v2.0
**作者**: Claude Code

---

## 一、项目概述

将原有的 SZTU-autologin 单文件脚本升级为功能完整、用户体验良好的校园网自动登录工具，支持宿舍区和教学区切换、可视化配置、自动重连和开机自启功能。

### 原有功能
- 单一硬编码的账号密码
- 固定的教学区认证（ac_id=1）
- 手动运行后单次登录

### 新增功能
- 账号密码通过配置文件便捷管理
- 宿舍区/教学区可切换
- 命令行交互式菜单（可视化面板）
- 定时检测网络状态，掉网自动重连
- Windows 任务计划开机自启
- 静默后台运行模式
- 完整的日志系统
- 单实例锁防止重复运行

---

## 二、技术栈

| 组件 | 版本 | 用途 |
|------|------|------|
| Python | 3.10+` | 主程序运行环境 |
| requests | - | HTTP 请求 |
| uv | - | 依赖管理和打包工具 |
| colorama | - | 命令行彩色输出 |

---

## 三、项目结构

```
SZTU-autologin/
├── main.py                 # 主入口，包含CLI菜单
├── build.py                 # 使用 uv 打包的脚本
├── core/
│   ├── __init__.py
│   ├── login_engine.py    # 登录核心逻辑（封装原加密算法）
│   ├── config.py          # 配置管理
│   ├── status.py          # 状态检测（是否已登录、网络状态）
│   ├── checker.py         # 后台自动检测重连
│   ├── scheduler.py       # 任务计划管理（开机自启）
│   ├── logger.py          # 日志系统
│   └── lock.py           # 单实例锁
├── utils/
│   ├── __init__.py
│   ├── network.py         # 网络工具（ping、检测WiFi）
│   └── crypto.py          # 加密算法（从原代码迁移）
├── config.json            # 用户配置（首次运行自动生成）
├── logs/                  # 日志目录
├── pyproject.toml         # uv 项目配置
├── README.md              # 使用说明
└── requirements.txt       # pip 兼容（可选）
```

---

## 四、配置文件设计

`config.json` 结构：

```json
{
  "account": {
    "username": "20xxxx@cucc",
    "password": "encrypted_or_plain_password",
    "isp": "cucc"
  },
  "network": {
    "area": "dormitory",  // "dormitory" | "teaching"
    "ac_id": "17",        // 宿舍区=17, 教学区=1
    "auto_reconnect": true,
    "check_interval": 300   // 秒，默认5分钟
  },
  "system": {
    "start_on_boot": false,
    "last_login": "2026-03-31 14:32:05",
    "version": "2.0.0"
  }
}
```

---

## 五、用户体验设计

### 5.1 首次使用引导

启动时检测 `config.json` 是否存在：

1. **不存在**：进入配置向导
   ```
   欢迎使用 SZTU 校园网自动登录工具！

   请按提示完成初始配置：

   [步骤 1/4] 输入学号
   > 输入学号: 20xxxx

   [步骤 2/4] 输入密码
   > 输入密码: ******

   [步骤 3/4] 选择运营商
     [1] 中国联通 (@cucc)
     [2] 中国移动 (@cmcc)
     [3] 中国电信 (@chinanet)
   > 选择: 1

   [步骤 4/4] 选择区域
     [1] 宿舍区 (北区宿舍)
     [2] 教学区 (教学区)
   > 选择: 1

   配置完成，正在测试登录...
   ✓ 登录成功！配置已保存。
   ```

2. **存在但版本不匹配**：提示用户升级配置格式

3. **配置损坏**：自动备份并重置

### 5.2 主菜单界面

```
══════════════════════════════════════════════════
    SZTU 校园网自动登录工具 v2.0
══════════════════════════════════════════════════
当前区域: 宿舍区 (ac_id=17)
登录状态: ● 已登录 (最后: 14:32:05)
自动检测: ● 已启用 (5分钟/次)
开机自启: ○ 已禁用
────────────────────────────────────────────────────────
[1] 修改账号密码
[2] 切换区域 (宿舍区/教学区)
[3] 手动登录 / 注销
[4] 开关自动检测重连
[5] 开关开机自启
[6] 查看日志
[7] 测试当前连接
[0] 退出 (后台运行)
────────────────────────────────────────────────────────
输入选项:
```

### 5.3 状态颜色编码
- 绿色 `●`：已启用/已登录
- 红色 `●`：已禁用/未登录
- 黄色 `●`：连接中/操作中

---

## 六、核心模块功能

### 6.1 core/login_engine.py

封装原有的登录逻辑，提供简洁接口：

```python
class LoginEngine:
    def __init__(self, config: dict):
        """初始化，根据配置设置 ac_id 等参数"""

    def login(self) -> LoginResult:
        """执行登录，返回结果对象"""
        # 1. 获取本机IP
        # 2. 获取 challenge token
        # 3. 执行加密计算
        # 4. 发送登录请求
        # 5. 解析响应

    def logout(self) -> LogoutResult:
        """执行注销"""
```

从原代码迁移：
- `get_xencode()` - 异或加密
- `get_base64()` - 自定义Base64编码
- `get_md5()` - HMAC-MD5加密
- `get_sha1()` - SHA-1校验和
- `get_token()` - 获取challenge token

### 6.2 core/config.py

配置文件管理：

```python
class ConfigManager:
    def load(self) -> dict: """加载配置"""
    def save(self, config: dict): """保存配置"""
    def validate(self, config: dict) -> bool: """验证配置完整性"""
    def migrate(self, old_version: str): """配置版本迁移"""
    def backup(self): """备份当前配置"""
```

### 6.3 core/status.py

状态检测：

```python
class StatusChecker:
    def is_logged_in(self) -> bool:
        """检测是否已登录（通过检测能否访问外网或返回特定响应）"""

    def is_network_available(self) -> bool:
        """检测网络是否可用"""

    def is_on_campus_network(self) -> bool:
        """检测是否连接到校园网认证网段"""

    def get_current_wifi(self) -> str:
        """获取当前连接的WiFi名称"""
```

### 6.4 core/checker.py

后台自动检测重连：

```python
class ConnectionChecker:
    def __init__(self, config: dict, login_engine: LoginEngine, logger: Logger):
        """初始化"""

    def start(self):
        """启动后台检测线程"""

    def stop(self):
        """停止检测"""

    def _check_loop(self):
        """检测主循环"""
        # while running:
        #   sleep(check_interval)
        #   if not is_logged_in():
        #       重试登录3次，间隔递增
        #       失败则暂停10分钟
```

连续失败处理：
- 失败次数 < 3：立即重试（间隔10s、30s、60s）
- 失败次数 >= 5：暂停检测10分钟
- 所有重试操作记录日志

### 6.5 core/scheduler.py

Windows 任务计划管理：

```python
class TaskScheduler:
    def create_task(self, exe_path: str) -> bool:
        """
        创建开机自启任务
        使用 schtasks 命令:
        schtasks /create /tn "SZTU-Autologin" /tr "exe_path --silent"
                  /sc onlogon /rl highest /f
        """

    def delete_task(self) -> bool:
        """删除任务: schtasks /delete /tn "SZTU-Autologin" """

    def task_exists(self) -> bool:
        """检查任务是否存在"""
```

任务配置：
- 任务名称：`SZTU-Autologin`
- 触发器：用户登录时 (`onlogon`)
- 运行权限：最高权限 (`/rl highest`)
- 运行参数：`--silent`（静默模式）

### 6.6 core/logger.py

日志系统：

```python
class Logger:
    def __init__(self, name: str, log_dir: str):
        """
        初始化
        - 按天轮转日志文件
        - 自动清理30天前的日志
        """

    def debug(self, msg: str): """DEBUG级别"""
    def info(self, msg: str): """INFO级别"""
    def warning(self, msg: str): """WARNING级别"""
    def error(self, msg: str, exc: Exception = None): """ERROR级别"""

    def get_recent_logs(self, lines: int = 50) -> list:
        """获取最近的日志"""
```

日志格式：`[2026-03-31 14:32:05] [INFO] [login_engine] 登录成功`

### 6.7 core/lock.py

单实例锁：

```python
class FileLock:
    def __init__(self, lock_file: str):
        """lock_file = "autologin.lock" """

    def acquire(self) -> bool:
        """获取锁，失败返回False"""

    def release(self):
        """释放锁"""

    def is_locked(self) -> bool:
        """检测是否已被锁定"""
```

---

## 七、工具模块

### 7.1 utils/network.py

```python
def ping(host: str, timeout: int = 3) -> bool:
    """检测主机是否可达"""

def get_local_ip() -> str:
    """获取本机IP（复用原代码逻辑）"""

def get_wifi_name() -> str:
    """获取当前WiFi名称（Windows: netsh wlan）"""
```

### 7.2 utils/crypto.py

从原 `autologin.py` 迁移所有加密算法：
- `force()`, `ordat()`, `sencode()`, `lencode()`
- `get_xencode()`, `get_base64()`
- `get_sha1()`, `get_md5()`

---

## 八、主程序流程 (main.py)

### 8.1 启动流程

```
main.py --silent 参数判断
├── 静默模式
│   ├── 检查单实例锁
│   ├── 加载配置
│   ├── 初始化各模块
│   ├── 启动自动检测线程
│   └── 保持运行
└── 交互模式
    ├── 检查单实例锁
    ├── 加载或创建配置
    ├── 初始化各模块
    ├── 显示主菜单
    └── 循环处理用户输入
```

### 8.2 菜单处理逻辑

| 选项 | 功能 |
|------|------|
| [1] | 修改账号密码（显示当前值，输入新值，验证保存） |
| [2] | 切换区域（宿舍区/教学区，更新ac_id） |
| [3] | 手动登录/注销（根据当前状态切换） |
| [4] | 开关自动检测（更新配置，重启检测线程） |
| [5] | 开关开机自启（调用TaskScheduler） |
| [6] | 查看日志（读取并显示日志） |
| [7] | 测试当前连接（显示详细状态信息） |
| [0] | 退出（检测后台运行选项） |

---

## 九、打包方案

使用 `uv` 打包，`build.py` 脚本：

```python
# build.py
import subprocess
import os

def build():
    # 1. 安装依赖
    subprocess.run(["uv", "add", "requests", "colorama"], check=True)

    # 2. 使用 pyinstaller 打包（需先安装）
    # pyinstaller -F -w --name SZTU-Autologin main.py
    subprocess.run(["pyinstaller", "-F", "-w", "--name", "SZTU-Autologin", "main.py"], check=True)

    # 3. 将dist目录中的exe复制到release目录
    # ...

if __name__ == "__main__":
    build()
```

打包步骤：
1. 使用 `uv` 安装依赖
2. 使用 `pyinstaller` 生成单文件exe（无控制台）
3. 生成release包，包含exe和README

---

## 十、错误处理与边界情况

| 场景 | 处理方式 |
|------|----------|
| 配置文件不存在 | 进入配置向导 |
| 配置文件损坏 | 备份并重置 |
| 单实例已存在 | 提示"程序已在运行"并退出 |
| 网络不可达 | 记录错误，等待下次检测 |
| 认证失败（密码错误） | 记录错误，暂停检测 |
| 服务器无响应 | 记录错误，等待下次检测 |
| 任务计划创建失败 | 提示权限问题 |
| 日志目录无写权限 | 降级到用户目录 |
| WiFi未连接 | 提示并等待 |

---

## 十一、安全性考虑

1. **密码存储**：当前为明文存储（与原项目一致），后续可考虑加密存储
2. **配置文件权限**：Windows 通过文件权限限制，Linux 使用 chmod 600
3. **日志敏感信息**：日志中不记录密码，只记录操作结果
4. **任务计划权限**：使用用户级别任务，非系统级

---

## 十二、版本管理

- **v2.0.0**：初始重构版本
  - 支持配置文件管理
  - 支持宿舍区/教学区切换
  - 命令行交互菜单
  - 自动检测重连
  - 开机自启

---

## 十三、发布说明

Release 包含：
- `SZTU-Autologin.exe` - 主程序
- `README.md` - 使用说明
- `CHANGELOG.md` - 更新日志

用户使用步骤：
1. 下载 `SZTU-Autologin.exe`
2. 首次运行，按向导配置
3. 如需开机自启，在菜单中启用

---

## 设计决策记录

1. **选择命令行菜单而非GUI**：用户明确要求C选项，最轻量化
2. **使用任务计划而非开机启动文件夹**：更稳定可靠
3. **检测间隔5分钟**：用户指定，平衡响应和负载
4. **使用JSON配置**：用户指定，简洁易读
5. **打包使用uv**：用户明确要求

---

## 后续优化方向

1. 密码加密存储
2. 跨平台支持（Linux/macOS）
3. 多账户支持
4. 网络流量统计
5. Telegram/企业微信通知功能
