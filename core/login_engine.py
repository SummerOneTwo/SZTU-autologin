import json
import re
import time
import requests

from utils import crypto, network

HTTP_TIMEOUT = 10

USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.26 Safari/537.36"

REQUEST_HEADERS = {"User-Agent": USER_AGENT}

GET_CHALLENGE_API = "http://172.19.0.5/cgi-bin/get_challenge"
SRUN_PORTAL_API = "http://172.19.0.5/cgi-bin/srun_portal"


class LoginResult:
    def __init__(self, success: bool, message: str = ""):
        self.success = success
        self.message = message


class LoginEngine:
    def __init__(self, config: dict, username: str = "", password: str = ""):
        self.config = config
        self.username = username
        self.password = password
        self.token = ""
        self.ip = ""
        self.ac_id = config["network"]["ac_id"]
        self.i = ""
        self.hmd5 = ""
        self.chksum = ""
        self.n = "200"
        self.login_type = "1"
        self.enc = "srun_bx1"

    def _init_getip(self):
        """获取本机IP"""
        self.ip = network.get_local_ip()

    def _get_token(self):
        """获取challenge token"""
        params = {
            "callback": "jQuery112406608265734960486_" + str(int(time.time() * 1000)),
            "username": self.username,
            "ip": self.ip,
            "_": int(time.time() * 1000),
        }
        resp = requests.get(
            GET_CHALLENGE_API,
            params=params,
            headers=REQUEST_HEADERS,
            timeout=HTTP_TIMEOUT,
        )
        match = re.search(r'"challenge":"(.*?)"', resp.text)
        if not match:
            raise ValueError("无法获取challenge token")
        self.token = match.group(1)
        return self.token

    def _get_info(self) -> str:
        info = {
            "username": self.username,
            "password": self.password,
            "ip": self.ip,
            "acid": self.ac_id,
            "enc_ver": self.enc,
        }
        return json.dumps(info, separators=(",", ":"))

    def _get_chksum(self) -> str:
        chkstr = self.token + self.username
        chkstr += self.token + self.hmd5
        chkstr += self.token + self.ac_id
        chkstr += self.token + self.ip
        chkstr += self.token + self.n
        chkstr += self.token + self.login_type
        chkstr += self.token + self.i
        return chkstr

    def _do_complex_work(self):
        self.i = self._get_info()
        self.i = "{SRBX1}" + crypto.get_base64(crypto.get_xencode(self.i, self.token))
        self.hmd5 = crypto.get_md5(self.password, self.token)
        self.chksum = crypto.get_sha1(self._get_chksum())

    def login(self) -> LoginResult:
        """执行登录"""
        try:
            self._init_getip()
            if not self.ip:
                return LoginResult(False, "无法获取本机IP地址")

            self._get_token()
            self._do_complex_work()

            params = {
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
                "type": self.login_type,
                "os": "windows 10",
                "name": "windows",
                "double_stack": 0,
                "_": int(time.time() * 1000),
            }
            resp = requests.get(
                SRUN_PORTAL_API,
                params=params,
                headers=REQUEST_HEADERS,
                timeout=HTTP_TIMEOUT,
            )

            if "ok" in resp.text:
                return LoginResult(True, "登录成功")
            else:
                match = re.search(r"\((.*?)\)", resp.text)
                if match:
                    error_data = match.group(1).replace("'", '"')
                    error_msg = json.loads(error_data)
                    return LoginResult(
                        False,
                        f"{error_msg.get('error', '')}: {error_msg.get('error_msg', '')}",
                    )
                return LoginResult(False, "登录失败：无法解析响应")
        except requests.exceptions.Timeout:
            return LoginResult(False, "请求超时")
        except requests.exceptions.ConnectionError:
            return LoginResult(False, "网络连接失败")
        except Exception as e:
            return LoginResult(False, str(e))
