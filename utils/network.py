import socket
import requests

USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.26 Safari/537.36"
REQUEST_HEADERS = {"User-Agent": USER_AGENT}
CAMPUS_GATEWAY = "172.19.0.5"
HTTP_TIMEOUT = 5


def get_local_ip() -> str:
    """获取本机IP地址"""
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as s:
            s.connect(("8.8.8.8", 80))
            return s.getsockname()[0]
    except Exception:
        return ""


def check_tcp_connect(host: str, port: int = 80, timeout: int = 3) -> bool:
    """检测主机TCP端口是否可达"""
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.settimeout(timeout)
            s.connect((host, port))
            return True
    except Exception:
        return False


def check_campus_network() -> bool:
    """检测是否连接到校园网"""
    try:
        resp = requests.get(
            f"http://{CAMPUS_GATEWAY}/",
            timeout=HTTP_TIMEOUT,
            headers=REQUEST_HEADERS,
            allow_redirects=False,
        )
        return resp.status_code in [200, 302, 301]
    except Exception:
        return False


def check_internet_access() -> bool:
    """检测是否已登录（能访问外网）"""
    try:
        resp = requests.get(
            "http://www.baidu.com",
            timeout=HTTP_TIMEOUT,
            headers=REQUEST_HEADERS,
            allow_redirects=False,
        )
        return resp.status_code == 200
    except Exception:
        return False
