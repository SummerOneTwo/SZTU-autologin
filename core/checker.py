import time
import threading
from datetime import datetime

from core.config import ConfigManager
from core.login_engine import LoginEngine
from core.status import StatusChecker
from core.logger import Logger

MAX_RETRIES = 3
MAX_CONSECUTIVE_FAILURES = 5
PAUSE_DURATION_ON_FAILURE = 600
PAUSE_CHECK_INTERVAL = 10
RETRY_INTERVALS = [10, 30, 60]


class ConnectionChecker:
    def __init__(
        self,
        config: dict,
        login_engine: LoginEngine,
        logger: Logger,
        cfg_mgr: ConfigManager,
    ):
        self.config = config
        self.login_engine = login_engine
        self.logger = logger
        self.cfg_mgr = cfg_mgr
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
                    time.sleep(PAUSE_CHECK_INTERVAL)
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
                        if self.consecutive_failures >= MAX_CONSECUTIVE_FAILURES:
                            self.logger.error("连续失败5次，暂停检测10分钟")
                            self.pause_until = datetime.fromtimestamp(
                                time.time() + PAUSE_DURATION_ON_FAILURE
                            )

            except Exception as e:
                self.logger.error("检测循环出错", e)

            time.sleep(check_interval)

    def _attempt_reconnect(self) -> bool:
        """尝试重连"""
        for attempt in range(MAX_RETRIES):
            self.logger.info(f"重连尝试 {attempt + 1}/{MAX_RETRIES}")
            result = self.login_engine.login()
            if result.success:
                self.logger.info("重连成功")
                self.cfg_mgr.update_last_login()
                return True
            else:
                self.logger.error(f"重连失败: {result.message}")
                if attempt < MAX_RETRIES - 1:
                    time.sleep(RETRY_INTERVALS[attempt])

        return False
