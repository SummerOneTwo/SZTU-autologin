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
