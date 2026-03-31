import copy
import json
import shutil
from pathlib import Path
from datetime import datetime


class ConfigManager:
    DEFAULT_CONFIG = {
        "account": {"username": "", "password": "", "isp": "cucc"},
        "network": {
            "area": "dormitory",
            "ac_id": "17",
            "auto_reconnect": True,
            "check_interval": 300,
        },
        "system": {"start_on_boot": False, "last_login": "", "version": "2.0.0"},
    }

    AREA_CONFIG = {
        "dormitory": {"ac_id": "17", "name": "宿舍区"},
        "teaching": {"ac_id": "1", "name": "教学区"},
    }

    ISP_CONFIG = {
        "cucc": {"suffix": "@cucc", "name": "中国联通"},
        "cmcc": {"suffix": "@cmcc", "name": "中国移动"},
        "chinanet": {"suffix": "@chinanet", "name": "中国电信"},
    }

    def __init__(self, config_file: str = "config.json"):
        self.config_file = Path(config_file)
        self._config = None

    def load(self) -> dict:
        """加载配置"""
        if not self.config_file.exists():
            self._config = copy.deepcopy(self.DEFAULT_CONFIG)
            return self._config

        try:
            with open(self.config_file, "r", encoding="utf-8") as f:
                self._config = json.load(f)
            self._validate_and_fix()
            return self._config
        except Exception:
            self.backup()
            self._config = copy.deepcopy(self.DEFAULT_CONFIG)
            return self._config

    def save(self):
        """保存配置"""
        if not self._config:
            self._config = copy.deepcopy(self.DEFAULT_CONFIG)

        try:
            with open(self.config_file, "w", encoding="utf-8") as f:
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
            try:
                backup_file = self.config_file.with_suffix(
                    f".json.backup.{datetime.now().strftime('%Y%m%d%H%M%S')}"
                )
                shutil.copy2(self.config_file, backup_file)
            except Exception:
                pass

    def update_last_login(self):
        """更新最后登录时间"""
        if self._config:
            self._config["system"]["last_login"] = datetime.now().strftime(
                "%Y-%m-%d %H:%M:%S"
            )
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

    def get_password(self) -> str:
        """获取密码"""
        if not self._config:
            self.load()
        return self._config["account"]["password"] or ""

    def set_password(self, password: str):
        """设置密码"""
        if not self._config:
            self.load()
        self._config["account"]["password"] = password
