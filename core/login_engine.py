import json
import re
import time
import requests

from utils import crypto, network


header = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.26 Safari/537.36"
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
        self.n = "200"
        self.type = "1"
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
            get_challenge_api, params=get_challenge_params, headers=header
        )
        self.token = re.search('"challenge":"(.*?)"', get_challenge_res.text).group(1)
        return self.token

    def _get_info(self):
        info_temp = {
            "username": self.username,
            "password": self.password,
            "ip": self.ip,
            "acid": self.ac_id,
            "enc_ver": self.enc,
        }
        i = re.sub("'", '"', str(info_temp))
        i = re.sub(" ", "", i)
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
                "callback": "jQuery11240645308969735664_"
                + str(int(time.time() * 1000)),
                "action": "login",
                "username": self.username,
                "password": "{MD5}" + self.hmd5,
                "ac_id": self.ac_id,
                "ip": self.ip,
                "chksum": self.chksum,
                "info": self.i,
                "n": self.n,
                "type": self.type,
                "os": "windows 10",
                "name": "windows",
                "double_stack": 0,
                "_": int(time.time() * 1000),
            }
            srun_portal_res = requests.get(
                srun_portal_api, params=srun_portal_params, headers=header
            )

            if "ok" in srun_portal_res.text:
                return LoginResult(True, "登录成功")
            else:
                match = re.search(r"\((.*?)\)", srun_portal_res.text)
                if match:
                    error_data = match.group(1).replace("'", '"')
                    error_msg = json.loads(error_data)
                    return LoginResult(
                        False,
                        f"{error_msg.get('error', '')}: {error_msg.get('error_msg', '')}",
                    )
                return LoginResult(False, "登录失败：无法解析响应")
        except Exception as e:
            return LoginResult(False, str(e))
