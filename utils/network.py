import socket
import requests

header = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.26 Safari/537.36"
}


def get_local_ip() -> str:
    """获取本机IP地址"""
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(("8.8.8.8", 80))
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
        resp = requests.get(
            "http://172.19.0.5/", timeout=5, headers=header, allow_redirects=False
        )
        return resp.status_code in [200, 302, 301]
    except Exception:
        return False


def check_internet_access() -> bool:
    """检测是否已登录（能访问外网）"""
    try:
        resp = requests.get(
            "http://www.baidu.com", timeout=5, headers=header, allow_redirects=False
        )
        if resp.status_code == 200:
            return True
        return False
    except Exception:
        return False
