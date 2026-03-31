from utils import network


class StatusChecker:
    def is_logged_in(self) -> bool:
        """检测是否已登录"""
        return network.check_internet_access()

    def is_on_campus_network(self) -> bool:
        """检测是否连接到校园网"""
        return network.check_campus_network()
