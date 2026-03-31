# SZTU 校园网自动登录工具 v2.0 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将单文件脚本重构为功能完整的校园网自动登录工具，支持宿舍区/教学区切换、配置文件管理、自动重连、开机自启等功能。

**Architecture:** 模块化重构，将原加密算法迁移到 `utils/crypto.py`，登录逻辑封装为 `LoginEngine`，配置、状态检测、自动检测、任务计划等作为独立模块，主程序通过命令行菜单交互。

**Tech Stack:** Python 3.10+, requests, colorama, uv

---

## 预备：项目结构与依赖初始化

**Files:**
- Create: `pyproject.toml`
- Create: `core/__init__.py`
- Create: `utils/__init__.py`
- Create: `logs/.gitkeep`

- [ ] **Step 1: 创建项目目录结构**

```bash
mkdir -p core utils logs
```

- [ ] **Step 2: 创建 pyproject.toml**

```toml
[project]
name = "sztu-autologin"
version = "2.0.0"
description = "SZTU 校园网自动登录工具"
requires-python = ">=3.10"
dependencies = [
    "requests>=2.32.3",
    "colorama>=0.4.6",
]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"
```

- [ ] **Step 3: 创建 core/__init__.py**

```python
from . import login_engine
from . import config
from . import status
from . import checker
from . import scheduler
from . import logger
from . import lock

__all__ = [
    "login_engine",
    "config",
    "status",
    "checker",
    "scheduler",
    "logger",
    "lock",
]
```

- [ ] **Step 4: 创建 utils/__init__.py**

```python
from . import network
from . import crypto

__all__ = ["network", "crypto"]
```

- [ ] **Step 5: 创建 logs/.gitkeep**

```
# 日志文件目录
```

---

## Task 1: 工具模块 - 加密算法迁移

**Files:**
- Create: `utils/crypto.py`

- [ ] **Step 1: 创建 utils/crypto.py 文件**

```python
# _*_ coding : utf-8 _*_
import math
import hashlib
import hmac

_PADCHAR = "="
_ALPHA = "LVoJPiCN2R8G90yg+hmFHuacZ1OWMnrsSTXkYpUq/3dlbfKwv6xztjI7DeBE45QA"


def force(msg):
    ret = []
    for w in msg:
        ret.append(ord(w))
    return bytes(ret)


def ordat(msg, idx):
    if len(msg) > idx:
        return ord(msg[idx])
    return 0


def sencode(msg, key):
    l = len(msg)
    pwd = []
    for i in range(0, l, 4):
        pwd.append(
            ordat(msg, i) | ordat(msg, i + 1) << 8 | ordat(msg, i + 2) << 16
            | ordat(msg, i + 3) << 24)
    if key:
        pwd.append(l)
    return pwd


def lencode(msg, key):
    l = len(msg)
    ll = (l - 1) << 2
    if key:
        m = msg[l - 1]
        if m < ll - 3 or m > ll:
            return
        ll = m
    for i in range(0, l):
        msg[i] = chr(msg[i] & 0xff) + chr(msg[i] >> 8 & 0xff) + chr(
            msg[i] >> 16 & 0xff) + chr(msg[i] >> 24 & 0xff)
    if key:
        return "".join(msg)[0:ll]
    return "".join(msg)


def get_xencode(msg, key):
    if msg == "":
        return ""
    pwd = sencode(msg, True)
    pwdk = sencode(key, False)
    if len(pwdk) < 4:
        pwdk = pwdk + [0] * (4 - len(pwdk))
    n = len(pwd) - 1
    z = pwd[n]
    y = pwd[0]
    c = 0x86014019 | 0x183639A0
    m = 0
    e = 0
    p = 0
    q = math.floor(6 + 52 / (n + 1))
    d = 0
    while 0 < q:
        d = d + c & (0x8CE0D9BF | 0x731F2640)
        e = d >> 2 & 3
        p = 0
        while p < n:
            y = pwd[p + 1]
            m = z >> 5 ^ y << 2
            m = m + ((y >> 3 ^ z << 4) ^ (d ^ y))
            m = m + (pwdk[(p & 3) ^ e] ^ z)
            pwd[p] = pwd[p] + m & (0xEFB8D130 | 0x10472ECF)
            z = pwd[p]
            p = p + 1
        y = pwd[0]
        m = z >> 5 ^ y << 2
        m = m + ((y >> 3 ^ z << 4) ^ (d ^ y))
        m = m + (pwdk[(p & 3) ^ e] ^ z)
        pwd[n] = pwd[n] + m & (0xBB390742 | 0x44C6F8BD)
        z = pwd[n]
        q = q - 1
    return lencode(pwd, False)


def _getbyte(s, i):
    x = ord(s[i])
    if x > 255:
        raise ValueError("INVALID_CHARACTER_ERR: DOM Exception 5")
    return x


def get_base64(s):
    i = 0
    b10 = 0
    x = []
    imax = len(s) - len(s) % 3
    if len(s) == 0:
        return s
    for i in range(0, imax, 3):
        b10 = (_getbyte(s, i) << 16) | (
            _getbyte(s, i + 1) << 8) | _getbyte(s, i + 2)
        x.append(_ALPHA[(b10 >> 18)])
        x.append(_ALPHA[((b10 >> 12) & 63)])
        x.append(_ALPHA[((b10 >> 6) & 63)])
        x.append(_ALPHA[(b10 & 63)])
    i = imax
    if len(s) - imax == 1:
        b10 = _getbyte(s, i) << 16
        x.append(_ALPHA[(b10 >> 18)] +
                 _ALPHA[((b10 >> 12) & 63)] + _PADCHAR + _PADCHAR)
    elif len(s) - imax == 2:
        b10 = _getbyte(s, i) << 16 | (_getbyte(s, i + 1) << 8
        x.append(_ALPHA[(b10 >> 18)] + _ALPHA[((b10 >> 12) & 63)
                                              ] + _ALPHA[((b10 >> 6) & 63)] + _PADCHAR)
    return "".join(x)


def get_sha1(value):
    return hashlib.sha1(value.encode()).hexdigest()


def get_md5(password, token):
    return hmac.new(token.encode(), password.encode(), hashlib.md5).hexdigest()
```

---

## Task 2: 工具模块 - 网络工具

**Files:**
- Create: `utils/network.py`

- [ ] **Step 1: 创建 utils/network.py**

```python
import socket
import requests

header = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.26 Safari/537.36'
}


def get_local_ip() -> str:
    """获取本机IP地址"""
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(('8.8.8.8', 80))
        ip = s.getsockname()[0]
        s.close()
        return ip
    except Exception:
        return ""


def ping_host(host: str, timeout: int = 3) -> bool:
    """检测主机是否可达"""
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.settimeout(timeout)
        s.connect((host, 80))
        s.close()
        return True
    except Exception:
        return False


def check_campus_network() -> bool:
    """检测是否连接到校园网"""
    try:
        resp = requests.get('http://172.19.0.5/', timeout=5, headers=header, allow_redirects=False)
        return resp.status_code in [200, 302, 301]
    except Exception:
        return False


def check_internet_access() -> bool:
    """检测是否已登录（能访问外网）"""
    try:
        resp = requests.get('http://www.baidu.com', timeout=5, headers=header, allow_redirects=False)
        if resp.status_code == 200:
            return True
        return False
    except Exception:
        return False
```

---

## Task 3: 核心模块 - 日志系统

**Files:**
- Create: `core/logger.py`

- [ ] **Step 1: 创建 core/logger.py**

```python
import os
import logging
from datetime import datetime
from pathlib import Path


class Logger:
    def __init__(self, name: str = "autologin", log_dir: str = "logs"):
        self.name = name
        self.log_dir = Path(log_dir)
        self.log_dir.mkdir(exist_ok=True)

        self._cleanup_old_logs()

        self.logger = logging.getLogger(name)
        self.logger.setLevel(logging.DEBUG)
        self.logger.handlers.clear()

        formatter = logging.Formatter(
            '[%(asctime)s] [%(levelname)s] [%(name)s] %(message)s',
            datefmt='%Y-%m-%d %H:%M:%S'
        )

        log_file = self.log_dir / f"{datetime.now().strftime('%Y-%m-%d')}.log"
        file_handler = logging.FileHandler(log_file, encoding='utf-8')
        file_handler.setLevel(logging.DEBUG)
        file_handler.setFormatter(formatter)
        self.logger.addHandler(file_handler)

        console_handler = logging.StreamHandler()
        console_handler.setLevel(logging.INFO)
        console_handler.setFormatter(formatter)
        self.logger.addHandler(console_handler)

    def _cleanup_old_logs(self):
        """清理30天前的日志"""
        try:
            now = datetime.now()
            for log_file in self.log_dir.glob("*.log"):
                mtime = datetime.fromtimestamp(log_file.stat().st_mtime)
                if (now - mtime).days > 30:
                    log_file.unlink()
        except Exception:
            pass

    def debug(self, msg: str):
        self.logger.debug(msg)

    def info(self, msg: str):
        self.logger.info(msg)

    def warning(self, msg: str):
        self.logger.warning(msg)

    def error(self, msg: str, exc: Exception = None):
        if exc:
            self.logger.error(msg, exc_info=True)
        else:
            self.logger.error(msg)

    def get_recent_logs(self, lines: int = 50) -> list:
        """获取最近的日志"""
        log_file = self.log_dir / f"{datetime.now().strftime('%Y-%m-%d')}.log"
        if not log_file.exists():
            return []
        try:
            with open(log_file, 'r', encoding='utf-8') as f:
                all_lines = f.readlines()
                return all_lines[-lines:] if len(all_lines) > lines else all_lines
        except Exception:
            return []
```

---

## Task 4: 核心模块 - 单实例锁

**Files:**
- Create: `core/lock.py`

- [ ] **Step 1: 创建 core/lock.py**

```python
import os
import sys
from pathlib import Path


class FileLock:
    def __init__(self, lock_file: str = "autologin.lock"):
        self.lock_file = Path(lock_file)
        self._locked = False

    def acquire(self) -> bool:
        """获取锁，失败返回False"""
        if self.lock_file.exists():
            try:
                with open(self.lock_file, 'r') as f:
                    pid = f.read().strip()
                    if pid:
                        if sys.platform == 'win32':
                            import ctypes
                            kernel32 = ctypes.windll.kernel32
                            try:
                                handle = kernel32.OpenProcess(1, False, int(pid))
                                if handle:
                                    kernel32.CloseHandle(handle)
                                    return False
                            except Exception:
                                pass
                        else:
                            try:
                                os.kill(int(pid), 0)
                                return False
                            except OSError:
                                pass
            except Exception:
                pass
        try:
            with open(self.lock_file, 'w') as f:
                f.write(str(os.getpid()))
            self._locked = True
            return True
        except Exception:
            return False

    def release(self):
        """释放锁"""
        if self._locked and self.lock_file.exists():
            try:
                self.lock_file.unlink()
            except Exception:
                pass
            self._locked = False

    def is_locked(self) -> bool:
        """检测是否已被锁定"""
        return self.lock_file.exists()

    def __enter__(self):
        if not self.acquire():
            raise RuntimeError("Another instance is already running")
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.release()
```

---

## Task 5: 核心模块 - 配置管理

**Files:**
- Create: `core/config.py`

- [ ] **Step 1: 创建 core/config.py**

```python
import json
import shutil
from pathlib import Path
from datetime import datetime


class ConfigManager:
    DEFAULT_CONFIG = {
        "account": {
            "username": "",
            "password": "",
            "isp": "cucc"
        },
        "network": {
            "area": "dormitory",
            "ac_id": "17",
            "auto_reconnect": True,
            "check_interval": 300
        },
        "system": {
            "start_on_boot": False,
            "last_login": "",
            "version": "2.0.0"
        }
    }

    AREA_CONFIG = {
        "dormitory": {"ac_id": "17", "name": "宿舍区"},
        "teaching": {"ac_id": "1", "name": "教学区"}
    }

    ISP_CONFIG = {
        "cucc": {"suffix": "@cucc", "name": "中国联通"},
        "cmcc": {"suffix": "@cmcc", "name": "中国移动"},
        "chinanet": {"suffix": "@chinanet", "name": "中国电信"}
    }

    def __init__(self, config_file: str = "config.json"):
        self.config_file = Path(config_file)
        self._config = None

    def load(self) -> dict:
        """加载配置"""
        if not self.config_file.exists():
            self._config = self.DEFAULT_CONFIG.copy()
            return self._config

        try:
            with open(self.config_file, 'r', encoding='utf-8') as f:
                self._config = json.load(f)
            self._validate_and_fix()
            return self._config
        except Exception:
            self.backup()
            self._config = self.DEFAULT_CONFIG.copy()
            return self._config

    def save(self, config: dict = None):
        """保存配置"""
        if config:
            self._config = config
        if not self._config:
            self._config = self.DEFAULT_CONFIG.copy()

        try:
            with open(self.config_file, 'w', encoding='utf-8') as f:
                json.dump(self._config, f, ensure_ascii=False, indent=2)
        except Exception as e:
            raise IOError(f"Failed to save config: {e}")

    def _validate_and_fix(self):
        """验证并修复配置"""
        for key, default in self.DEFAULT_CONFIG.items():
            if key not in self._config:
                self._config[key] = default
            elif isinstance(default, dict):
                for subkey, subdefault in default.items():
                    if subkey not in self._config[key]:
                        self._config[key][subkey] = subdefault

        area = self._config["network"]["area"]
        if area not in self.AREA_CONFIG:
            area = "dormitory"
            self._config["network"]["area"] = area
        self._config["network"]["ac_id"] = self.AREA_CONFIG[area]["ac_id"]

    def validate(self, config: dict) -> bool:
        """验证配置完整性"""
        if not config.get("account", {}).get("username"):
            return False
        if not config.get("account", {}).get("password"):
            return False
        if config.get("network", {}).get("area") not in self.AREA_CONFIG:
            return False
        return True

    def backup(self):
        """备份当前配置"""
        if self.config_file.exists():
            backup_file = self.config_file.with_suffix(f'.json.backup.{datetime.now().strftime("%Y%m%d%H%M%S")}')
            shutil.copy2(self.config_file, backup_file)

    def update_last_login(self):
        """更新最后登录时间"""
        if self._config:
            self._config["system"]["last_login"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            self.save()

    def get_full_username(self) -> str:
        """获取完整用户名（含运营商后缀）"""
        if not self._config:
            self.load()
        username = self._config["account"]["username"]
        isp = self._config["account"]["isp"]
        if "@" in username:
            return username
        suffix = self.ISP_CONFIG.get(isp, {}).get("suffix", "")
        return f"{username}{suffix}"
```

---

## Task 6: 核心模块 - 登录引擎

**Files:**
- Create: `core/login_engine.py`

- [ ] **Step 1: 创建 core/login_engine.py**

```python
import re
import time
import requests
from urllib.parse import quote

from utils import crypto, network


header = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.26 Safari/537.36'
}

get_challenge_api = "http://172.19.0.5/cgi-bin/get_challenge"
srun_portal_api = "http://172.19.0.5/cgi-bin/srun_portal"


class LoginResult:
    def __init__(self, success: bool, message: str = ""):
        self.success = success
        self.message = message


class LoginEngine:
    def __init__(self, config: dict):
        self.config = config
        self.token = ""
        self.ip = ""
        self.username = ""
        self.password = ""
        self.ac_id = config["network"]["ac_id"]
        self.i = ""
        self.hmd5 = ""
        self.chksum = ""
        self.n = '200'
        self.type = '1'
        self.enc = "srun_bx1"

    def _init_getip(self):
        """获取本机IP"""
        self.ip = network.get_local_ip()

    def _get_token(self):
        """获取challenge token"""
        get_challenge_params = {
            "callback": "jQuery112406608265734960486_" + str(int(time.time() * 1000)),
            "username": self.username,
            "ip": self.ip,
            "_": int(time.time() * 1000),
        }
        get_challenge_res = requests.get(
            get_challenge_api, params=get_challenge_params, headers=header)
        self.token = re.search('"challenge":"(.*?)"', get_challenge_res.text).group(1)
        return self.token

    def _get_info(self):
        info_temp = {
            "username": self.username,
            "password": self.password,
            "ip": self.ip,
            "acid": self.ac_id,
            "enc_ver": self.enc
        }
        i = re.sub("'", '"', str(info_temp))
        i = re.sub(" ", '', i)
        return i

    def _get_chksum(self):
        chkstr = self.token + self.username
        chkstr += self.token + self.hmd5
        chkstr += self.token + self.ac_id
        chkstr += self.token + self.ip
        chkstr += self.token + self.n
        chkstr += self.token + self.type
        chkstr += self.token + self.i
        return chkstr

    def _do_complex_work(self):
        self.i = self._get_info()
        self.i = "{SRBX1}" + crypto.get_base64(crypto.get_xencode(self.i, self.token))
        self.hmd5 = crypto.get_md5(self.password, self.token)
        self.chksum = crypto.get_sha1(self._get_chksum())

    def login(self) -> LoginResult:
        """执行登录"""
        from core.config import ConfigManager

        cfg_mgr = ConfigManager()
        self.username = cfg_mgr.get_full_username()
        self.password = self.config["account"]["password"]
        self.ac_id = self.config["network"]["ac_id"]

        try:
            self._init_getip()
            if not self.ip:
                return LoginResult(False, "无法获取本机IP地址")

            self._get_token()
            self._do_complex_work()

            srun_portal_params = {
                'callback': 'jQuery11240645308969735664_' + str(int(time.time() * 1000)),
                'action': 'login',
                'username': self.username,
                'password': '{MD5}' + self.hmd5,
                'ac_id': self.ac_id,
                'ip': self.ip,
                'chksum': self.chksum,
                'info': self.i,
                'n': self.n,
                'type': self.type,
                'os': 'windows 10',
                'name': 'windows',
                'double_stack': 0,
                '_': int(time.time() * 1000),
            }
            srun_portal_res = requests.get(
                srun_portal_api, params=srun_portal_params, headers=header)

            if 'ok' in srun_portal_res.text:
                return LoginResult(True, "登录成功")
            else:
                error_msg = eval(re.search('\((.*?)\)', srun_portal_res.text).group(1))
                return LoginResult(False, f"{error_msg.get('error', '')}: {error_msg.get('error_msg', '')}")
        except Exception as e:
            return LoginResult(False, str(e))
```

---

## Task 7: 核心模块 - 状态检测

**Files:**
- Create: `core/status.py`

- [ ] **Step 1: 创建 core/status.py**

```python
from utils import network


class StatusChecker:
    def __init__(self):
        pass

    def is_logged_in(self) -> bool:
        """检测是否已登录"""
        return network.check_internet_access()

    def is_network_available(self) -> bool:
        """检测网络是否可用"""
        return network.ping_host("172.19.0.5", timeout=3) or self.is_on_campus_network()

    def is_on_campus_network(self) -> bool:
        """检测是否连接到校园网"""
        return network.check_campus_network()
```

---

## Task 8: 核心模块 - 任务计划管理

**Files:**
- Create: `core/scheduler.py`

- [ ] **Step 1: 创建 core/scheduler.py**

```python
import os
import sys
import subprocess


class TaskScheduler:
    TASK_NAME = "SZTU-Autologin"

    @classmethod
    def create_task(cls, exe_path: str = None) -> bool:
        """创建开机自启任务"""
        if not exe_path:
            exe_path = sys.executable
            if exe_path.endswith("python.exe") or exe_path.endswith("pythonw.exe"):
                script_path = os.path.abspath(sys.argv[0])
                exe_path = f'"{exe_path}" "{script_path}"'
            else:
                exe_path = f'"{exe_path}"'

        cmd = [
            "schtasks", "/create", "/tn", cls.TASK_NAME,
            "/tr", f"{exe_path} --silent",
            "/sc", "onlogon", "/rl", "highest", "/f"
        ]

        try:
            result = subprocess.run(cmd, capture_output=True, text=True, shell=True)
            return result.returncode == 0
        except Exception:
            return False

    @classmethod
    def delete_task(cls) -> bool:
        """删除任务"""
        cmd = ["schtasks", "/delete", "/tn", cls.TASK_NAME, "/f"]
        try:
            result = subprocess.run(cmd, capture_output=True, text=True, shell=True)
            return result.returncode == 0
        except Exception:
            return False

    @classmethod
    def task_exists(cls) -> bool:
        """检查任务是否存在"""
        cmd = ["schtasks", "/query", "/tn", cls.TASK_NAME]
        try:
            result = subprocess.run(cmd, capture_output=True, text=True, shell=True)
            return result.returncode == 0
        except Exception:
            return False
```

---

## Task 9: 核心模块 - 自动检测重连

**Files:**
- Create: `core/checker.py`

- [ ] **Step 1: 创建 core/checker.py**

```python
import time
import threading
from datetime import datetime

from core.login_engine import LoginEngine
from core.status import StatusChecker
from core.logger import Logger


class ConnectionChecker:
    def __init__(self, config: dict, login_engine: LoginEngine, logger: Logger):
        self.config = config
        self.login_engine = login_engine
        self.logger = logger
        self.status_checker = StatusChecker()
        self.running = False
        self.thread = None
        self.consecutive_failures = 0
        self.pause_until = None

    def start(self):
        """启动后台检测线程"""
        if self.running:
            return
        self.running = True
        self.thread = threading.Thread(target=self._check_loop, daemon=True)
        self.thread.start()
        self.logger.info("自动检测已启动")

    def stop(self):
        """停止检测"""
        self.running = False
        if self.thread and self.thread.is_alive():
            self.thread.join(timeout=2)
        self.logger.info("自动检测已停止")

    def _check_loop(self):
        """检测主循环"""
        check_interval = self.config["network"]["check_interval"]

        while self.running:
            try:
                if self.pause_until and datetime.now() < self.pause_until:
                    time.sleep(10)
                    continue
                self.pause_until = None

                if not self.status_checker.is_on_campus_network():
                    self.logger.debug("未连接到校园网，跳过检测")
                    time.sleep(check_interval)
                    continue

                if not self.status_checker.is_logged_in():
                    self.logger.warning("检测到网络断开，尝试重连...")
                    if self._attempt_reconnect():
                        self.consecutive_failures = 0
                    else:
                        self.consecutive_failures += 1
                        if self.consecutive_failures >= 5:
                            self.logger.error("连续失败5次，暂停检测10分钟")
                            self.pause_until = datetime.fromtimestamp(
                                time.time() + 600
                            )

            except Exception as e:
                self.logger.error("检测循环出错", e)

            time.sleep(check_interval)

    def _attempt_reconnect(self) -> bool:
        """尝试重连"""
        retry_intervals = [10, 30, 60]
        max_retries = 3

        for attempt in range(max_retries):
            self.logger.info(f"重连尝试 {attempt + 1}/{max_retries}")
            result = self.login_engine.login()
            if result.success:
                self.logger.info("重连成功")
                from core.config import ConfigManager
                ConfigManager().update_last_login()
                return True
            else:
                self.logger.error(f"重连失败: {result.message}")
                if attempt < max_retries - 1:
                    time.sleep(retry_intervals[attempt])

        return False
```

---

## Task 10: 主程序 - 命令行菜单

**Files:**
- Create: `main.py`
- Modify: 原 `autologin.py` 保留为备份

- [ ] **Step 1: 创建 main.py**

```python
#!/usr/bin/env python3
# _*_ coding : utf-8 _*_
import sys
import os
import time
import argparse
from pathlib import Path

from colorama import init, Fore, Style

from core.config import ConfigManager
from core.login_engine import LoginEngine, LoginResult
from core.status import StatusChecker
from core.checker import ConnectionChecker
from core.scheduler import TaskScheduler
from core.logger import Logger
from core.lock import FileLock


init(autoreset=True)


class AutologinApp:
    def __init__(self):
        self.cfg_mgr = ConfigManager()
        self.config = self.cfg_mgr.load()
        self.logger = Logger()
        self.status_checker = StatusChecker()
        self.login_engine = None
        self.checker = None
        self.running = True

    def setup_login_engine(self):
        if not self.login_engine:
            self.login_engine = LoginEngine(self.config)
        return self.login_engine

    def setup_checker(self):
        if not self.checker:
            self.checker = ConnectionChecker(self.config, self.setup_login_engine(), self.logger)
        return self.checker

    def show_status(self):
        area_name = ConfigManager.AREA_CONFIG[self.config["network"]["area"]]["name"]
        ac_id = self.config["network"]["ac_id"]
        is_logged_in = self.status_checker.is_logged_in()
        auto_reconnect = self.config["network"]["auto_reconnect"]
        start_on_boot = TaskScheduler.task_exists()
        last_login = self.config["system"]["last_login"] or "从未"

        print(f"\n{Fore.CYAN}{'='*55}{Style.RESET_ALL}")
        print(f"    {Fore.CYAN}SZTU 校园网自动登录工具 v2.0{Style.RESET_ALL}")
        print(f"{Fore.CYAN}{'='*55}{Style.RESET_ALL}")

        status_color = Fore.GREEN if is_logged_in else Fore.RED
        status_dot = f"{status_color}●{Style.RESET_ALL}" if is_logged_in else f"{Fore.RED}●{Style.RESET_ALL}"
        auto_color = Fore.GREEN if auto_reconnect else Fore.RED
        auto_dot = f"{auto_color}●{Style.RESET_ALL}" if auto_reconnect else f"{Fore.RED}○{Style.RESET_ALL}"
        boot_color = Fore.GREEN if start_on_boot else Fore.RED
        boot_dot = f"{boot_color}●{Style.RESET_ALL}" if start_on_boot else f"{Fore.RED}○{Style.RESET_ALL}"

        print(f"当前区域: {area_name} (ac_id={ac_id})")
        print(f"登录状态: {status_dot} {'已登录' if is_logged_in else '未登录'} (最后: {last_login})")
        print(f"自动检测: {auto_dot} {'已启用' if auto_reconnect else '已禁用'} ({self.config['network']['check_interval']}秒/次)")
        print(f"开机自启: {boot_dot} {'已启用' if start_on_boot else '已禁用'}")
        print(f"{Fore.CYAN}{'-'*55}{Style.RESET_ALL}")

    def show_menu(self):
        print("\n[1] 修改账号密码")
        print("[2] 切换区域 (宿舍区/教学区)")
        print("[3] 手动登录 / 注销")
        print("[4] 开关自动检测重连")
        print("[5] 开关开机自启")
        print("[6] 查看日志")
        print("[7] 测试当前连接")
        print("[0] 退出")
        print(f"{Fore.CYAN}{'-'*55}{Style.RESET_ALL}")

    def config_wizard(self):
        print(f"\n{Fore.GREEN}欢迎使用 SZTU 校园网自动登录工具！{Style.RESET_ALL}")
        print("\n请按提示完成初始配置：\n")

        print(f"{Fore.YELLOW}[步骤 1/4] 输入学号{Style.RESET_ALL}")
        student_id = input("> 输入学号: ").strip()

        print(f"\n{Fore.YELLOW}[步骤 2/4] 输入密码{Style.RESET_ALL}")
        password = input("> 输入密码: ").strip()

        print(f"\n{Fore.YELLOW}[步骤 3/4] 选择运营商{Style.RESET_ALL}")
        print("  [1] 中国联通 (@cucc)")
        print("  [2] 中国移动 (@cmcc)")
        print("  [3] 中国电信 (@chinanet)")
        isp_choice = input("> 选择: ").strip()
        isp_map = {"1": "cucc", "2": "cmcc", "3": "chinanet"}
        isp = isp_map.get(isp_choice, "cucc")

        print(f"\n{Fore.YELLOW}[步骤 4/4] 选择区域{Style.RESET_ALL}")
        print("  [1] 宿舍区 (北区宿舍)")
        print("  [2] 教学区 (教学区)")
        area_choice = input("> 选择: ").strip()
        area = "dormitory" if area_choice == "1" else "teaching"

        self.config["account"]["username"] = student_id
        self.config["account"]["password"] = password
        self.config["account"]["isp"] = isp
        self.config["network"]["area"] = area
        self.config["network"]["ac_id"] = ConfigManager.AREA_CONFIG[area]["ac_id"]
        self.cfg_mgr.save(self.config)

        print(f"\n{Fore.GREEN}配置完成，正在测试登录...{Style.RESET_ALL}")
        result = self.setup_login_engine().login()
        if result.success:
            print(f"{Fore.GREEN}✓ 登录成功！配置已保存。{Style.RESET_ALL}")
            self.cfg_mgr.update_last_login()
        else:
            print(f"{Fore.RED}✗ 登录失败: {result.message}{Style.RESET_ALL}")

    def modify_account(self):
        print(f"\n{Fore.YELLOW}修改账号密码{Style.RESET_ALL}")
        print(f"当前用户名: {self.config['account']['username']}")
        print("(直接回车保持原值)")

        new_username = input("> 输入新学号: ").strip()
        if new_username:
            self.config["account"]["username"] = new_username

        new_password = input("> 输入新密码: ").strip()
        if new_password:
            self.config["account"]["password"] = new_password

        print("\n选择运营商:")
        print("  [1] 中国联通 (@cucc)")
        print("  [2] 中国移动 (@cmcc)")
        print("  [3] 中国电信 (@chinanet)")
        isp_choice = input("> 选择 (回车保持当前): ").strip()
        if isp_choice in ["1", "2", "3"]:
            isp_map = {"1": "cucc", "2": "cmcc", "3": "chinanet"}
            self.config["account"]["isp"] = isp_map[isp_choice]

        self.cfg_mgr.save(self.config)
        print(f"{Fore.GREEN}账号密码已更新{Style.RESET_ALL}")

    def switch_area(self):
        print(f"\n{Fore.YELLOW}切换区域{Style.RESET_ALL}")
        print(f"当前区域: {ConfigManager.AREA_CONFIG[self.config['network']['area']]['name']}")
        print("\n  [1] 宿舍区 (北区宿舍, ac_id=17)")
        print("  [2] 教学区 (教学区, ac_id=1)")
        choice = input("> 选择: ").strip()

        if choice == "1":
            self.config["network"]["area"] = "dormitory"
            self.config["network"]["ac_id"] = "17"
        elif choice == "2":
            self.config["network"]["area"] = "teaching"
            self.config["network"]["ac_id"] = "1"
        else:
            return

        self.cfg_mgr.save(self.config)
        print(f"{Fore.GREEN}已切换到 {ConfigManager.AREA_CONFIG[self.config['network']['area']]['name']}{Style.RESET_ALL}")

    def manual_login(self):
        is_logged_in = self.status_checker.is_logged_in()
        if is_logged_in:
            print(f"\n{Fore.YELLOW}当前已登录，确定要重新登录吗？{Style.RESET_ALL}")
            confirm = input("(y/n): ").strip().lower()
            if confirm != 'y':
                return

        print(f"\n{Fore.YELLOW}正在登录...{Style.RESET_ALL}")
        result = self.setup_login_engine().login()
        if result.success:
            print(f"{Fore.GREEN}✓ {result.message}{Style.RESET_ALL}")
            self.cfg_mgr.update_last_login()
        else:
            print(f"{Fore.RED}✗ {result.message}{Style.RESET_ALL}")

    def toggle_auto_reconnect(self):
        self.config["network"]["auto_reconnect"] = not self.config["network"]["auto_reconnect"]
        self.cfg_mgr.save(self.config)
        status = "已启用" if self.config["network"]["auto_reconnect"] else "已禁用"
        print(f"\n{Fore.GREEN}自动检测重连 {status}{Style.RESET_ALL}")

        if self.config["network"]["auto_reconnect"] and self.checker:
            self.setup_checker().start()
        elif self.checker:
            self.checker.stop()

    def toggle_start_on_boot(self):
        current = TaskScheduler.task_exists()
        if current:
            print(f"\n{Fore.YELLOW}当前开机自启已启用，确定要关闭吗？{Style.RESET_ALL}")
            confirm = input("(y/n): ").strip().lower()
            if confirm != 'y':
                return
            if TaskScheduler.delete_task():
                self.config["system"]["start_on_boot"] = False
                self.cfg_mgr.save(self.config)
                print(f"{Fore.GREEN}开机自启已关闭{Style.RESET_ALL}")
            else:
                print(f"{Fore.RED}关闭失败{Style.RESET_ALL}")
        else:
            print(f"\n{Fore.YELLOW}确定要启用开机自启吗？{Style.RESET_ALL}")
            confirm = input("(y/n): ").strip().lower()
            if confirm != 'y':
                return
            if TaskScheduler.create_task():
                self.config["system"]["start_on_boot"] = True
                self.cfg_mgr.save(self.config)
                print(f"{Fore.GREEN}开机自启已启用{Style.RESET_ALL}")
            else:
                print(f"{Fore.RED}启用失败，请以管理员身份运行{Style.RESET_ALL}")

    def view_logs(self):
        print(f"\n{Fore.YELLOW}最近日志:{Style.RESET_ALL}")
        logs = self.logger.get_recent_logs(50)
        if not logs:
            print("(无日志)")
            return
        for line in logs[-30:]:
            print(line.rstrip())

    def test_connection(self):
        print(f"\n{Fore.YELLOW}测试连接状态...{Style.RESET_ALL}")
        on_campus = self.status_checker.is_on_campus_network()
        print(f"校园网可达: {Fore.GREEN if on_campus else Fore.RED}{'是' if on_campus else '否'}{Style.RESET_ALL}")
        is_logged_in = self.status_checker.is_logged_in()
        print(f"外网访问: {Fore.GREEN if is_logged_in else Fore.RED}{'正常' if is_logged_in else '未登录'}{Style.RESET_ALL}")

    def run_interactive(self):
        if not self.cfg_mgr.validate(self.config):
            self.config_wizard()
            if not self.cfg_mgr.validate(self.config):
                print(f"{Fore.RED}配置不完整，程序将退出{Style.RESET_ALL}")
                return

        if self.config["network"]["auto_reconnect"]:
            self.setup_checker().start()

        while self.running:
            self.show_status()
            self.show_menu()
            choice = input("输入选项: ").strip()

            if choice == "1":
                self.modify_account()
            elif choice == "2":
                self.switch_area()
            elif choice == "3":
                self.manual_login()
            elif choice == "4":
                self.toggle_auto_reconnect()
            elif choice == "5":
                self.toggle_start_on_boot()
            elif choice == "6":
                self.view_logs()
            elif choice == "7":
                self.test_connection()
            elif choice == "0":
                self.running = False
                if self.checker:
                    self.checker.stop()
                print(f"\n{Fore.GREEN}再见！{Style.RESET_ALL}")
            else:
                print(f"{Fore.RED}无效选项{Style.RESET_ALL}")

            if choice not in ["0"]:
                input("\n按回车继续...")

    def run_silent(self):
        if not self.cfg_mgr.validate(self.config):
            self.logger.error("配置不完整，静默模式无法运行")
            return

        self.logger.info("静默模式启动")
        if self.config["network"]["auto_reconnect"]:
            self.setup_checker().start()

        try:
            while True:
                time.sleep(1)
        except KeyboardInterrupt:
            pass

        if self.checker:
            self.checker.stop()


def main():
    parser = argparse.ArgumentParser(description="SZTU 校园网自动登录工具")
    parser.add_argument("--silent", action="store_true", help="静默模式运行")
    args = parser.parse_args()

    lock = FileLock()
    if not lock.acquire():
        print(f"{Fore.RED}程序已在运行中{Style.RESET_ALL}")
        return

    try:
        app = AutologinApp()
        if args.silent:
            app.run_silent()
        else:
            app.run_interactive()
    finally:
        lock.release()


if __name__ == "__main__":
    main()
```

---

## Task 11: 打包脚本与更新文档

**Files:**
- Create: `build.py`
- Update: `README.md`
- Create: `CHANGELOG.md`

- [ ] **Step 1: 创建 build.py**

```python
#!/usr/bin/env python3
import subprocess
import os
import shutil
from pathlib import Path


def build():
    print("安装依赖...")
    subprocess.run(["uv", "add", "requests", "colorama"], check=True)

    print("\n使用 PyInstaller 打包...")
    subprocess.run([
        "uv", "run", "pyinstaller",
        "-F", "-w",
        "--name", "SZTU-Autologin",
        "main.py"
    ], check=True)

    dist_path = Path("dist")
    if dist_path.exists():
        print(f"\n打包完成！可执行文件位于: {dist_path.absolute()}")
        exe_file = dist_path / "SZTU-Autologin.exe"
        if exe_file.exists():
            print(f"可执行文件: {exe_file}")


if __name__ == "__main__":
    build()
```

- [ ] **Step 2: 更新 README.md**

```markdown
# SZTU-autologin
SZTU深圳技术大学校园网自动登录脚本 v2.0

觉得好用不妨点个star😋

## 功能特性

- ✅ 支持宿舍区和教学区切换
- ✅ 账号密码便捷配置
- ✅ 命令行交互式菜单
- ✅ 定时检测网络状态，掉网自动重连
- ✅ Windows任务计划开机自启
- ✅ 后台静默运行
- ✅ 完整的日志系统

## 快速开始

### 方式一：使用发布版本（推荐）

1. 从 [Releases](../../releases) 下载最新的 `SZTU-Autologin.exe`
2. 双击运行，按提示完成配置
3. 在菜单中启用开机自启

### 方式二：从源码运行

1. 克隆仓库，使用 `uv` 管理依赖：

```bash
uv sync
uv run main.py
```

### 方式三：打包为可执行文件

```bash
uv add pyinstaller
uv run pyinstaller -F -w --name SZTU-Autologin main.py
```

## 使用说明

首次运行会进入配置向导，按提示输入：
1. 学号
2. 密码
3. 选择运营商（联通/移动/电信）
4. 选择区域（宿舍区/教学区）

主菜单选项：
- [1] 修改账号密码
- [2] 切换区域
- [3] 手动登录/注销
- [4] 开关自动检测重连
- [5] 开关开机自启
- [6] 查看日志
- [7] 测试当前连接
- [0] 退出

## 更新日志

详见 [CHANGELOG.md](./CHANGELOG.md)

## 致谢

本项目基于 [shadow-aaa/SZTU-autologin](https://github.com/shadow-aaa/SZTU-autologin) 重构而来。
```

- [ ] **Step 3: 创建 CHANGELOG.md**

```markdown
# Changelog

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
```

---

## Task 12: 更新 .gitignore 和 requirements.txt

**Files:**
- Update: `.gitignore`
- Update: `requirements.txt`

- [ ] **Step 1: 更新 .gitignore**

```
*.pyc
__pycache__/
*.py[cod]
*$py.class
*.so
.Python
build/
develop-eggs/
dist/
downloads/
eggs/
.eggs/
lib/
lib64/
parts/
sdist/
var/
wheels/
*.egg-info/
.installed.cfg
*.egg
.venv/
env/
venv/
ENV/
config.json
*.log
logs/
autologin.lock
*.spec
```

- [ ] **Step 2: 更新 requirements.txt**

```
requests>=2.32.3
colorama>=0.4.6
```

---

## Task 13: 集成测试

**Files:**
- Run: 手动测试

- [ ] **Step 1: 运行程序测试**

```bash
uv add requests colorama
uv run main.py
```

检查：
- 配置向导能正常工作
- 主菜单显示正常
- 状态显示正确
- 所有菜单项功能正常

---

## 附录：设计文档与规范

- 项目结构与设计文档一致
- 所有模块职责单一明确
- 配置文件格式符合规范
- 日志记录完整清晰
- 用户交互友好直观
