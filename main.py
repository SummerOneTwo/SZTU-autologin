#!/usr/bin/env python3
# _*_ coding : utf-8 _*_
import time
import argparse

from colorama import init, Fore, Style

from core.config import ConfigManager
from core.login_engine import LoginEngine
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
            self.checker = ConnectionChecker(
                self.config, self.setup_login_engine(), self.logger
            )
        return self.checker

    def show_status(self):
        area_name = ConfigManager.AREA_CONFIG[self.config["network"]["area"]]["name"]
        ac_id = self.config["network"]["ac_id"]
        is_logged_in = self.status_checker.is_logged_in()
        auto_reconnect = self.config["network"]["auto_reconnect"]
        start_on_boot = TaskScheduler.task_exists()
        last_login = self.config["system"]["last_login"] or "从未"

        print(f"\n{Fore.CYAN}{'=' * 55}{Style.RESET_ALL}")
        print(f"    {Fore.CYAN}SZTU 校园网自动登录工具 v2.0{Style.RESET_ALL}")
        print(f"{Fore.CYAN}{'=' * 55}{Style.RESET_ALL}")

        status_color = Fore.GREEN if is_logged_in else Fore.RED
        status_dot = (
            f"{status_color}●{Style.RESET_ALL}"
            if is_logged_in
            else f"{Fore.RED}●{Style.RESET_ALL}"
        )
        auto_color = Fore.GREEN if auto_reconnect else Fore.RED
        auto_dot = (
            f"{auto_color}●{Style.RESET_ALL}"
            if auto_reconnect
            else f"{Fore.RED}○{Style.RESET_ALL}"
        )
        boot_color = Fore.GREEN if start_on_boot else Fore.RED
        boot_dot = (
            f"{boot_color}●{Style.RESET_ALL}"
            if start_on_boot
            else f"{Fore.RED}○{Style.RESET_ALL}"
        )

        print(f"当前区域: {area_name} (ac_id={ac_id})")
        print(
            f"登录状态: {status_dot} {'已登录' if is_logged_in else '未登录'} (最后: {last_login})"
        )
        print(
            f"自动检测: {auto_dot} {'已启用' if auto_reconnect else '已禁用'} ({self.config['network']['check_interval']}秒/次)"
        )
        print(f"开机自启: {boot_dot} {'已启用' if start_on_boot else '已禁用'}")
        print(f"{Fore.CYAN}{'-' * 55}{Style.RESET_ALL}")

    def show_menu(self):
        print("\n[1] 修改账号密码")
        print("[2] 切换区域 (宿舍区/教学区)")
        print("[3] 手动登录")
        print("[4] 开关自动检测重连")
        print("[5] 开关开机自启")
        print("[6] 查看日志")
        print("[7] 测试当前连接")
        print("[0] 退出")
        print(f"{Fore.CYAN}{'-' * 55}{Style.RESET_ALL}")

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
        print(
            f"当前区域: {ConfigManager.AREA_CONFIG[self.config['network']['area']]['name']}"
        )
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
        print(
            f"{Fore.GREEN}已切换到 {ConfigManager.AREA_CONFIG[self.config['network']['area']]['name']}{Style.RESET_ALL}"
        )

    def manual_login(self):
        is_logged_in = self.status_checker.is_logged_in()
        if is_logged_in:
            print(f"\n{Fore.YELLOW}当前已登录，确定要重新登录吗？{Style.RESET_ALL}")
            confirm = input("(y/n): ").strip().lower()
            if confirm != "y":
                return

        print(f"\n{Fore.YELLOW}正在登录...{Style.RESET_ALL}")
        result = self.setup_login_engine().login()
        if result.success:
            print(f"{Fore.GREEN}✓ {result.message}{Style.RESET_ALL}")
            self.cfg_mgr.update_last_login()
        else:
            print(f"{Fore.RED}✗ {result.message}{Style.RESET_ALL}")

    def toggle_auto_reconnect(self):
        self.config["network"]["auto_reconnect"] = not self.config["network"][
            "auto_reconnect"
        ]
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
            if confirm != "y":
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
            if confirm != "y":
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
        print(
            f"校园网可达: {Fore.GREEN if on_campus else Fore.RED}{'是' if on_campus else '否'}{Style.RESET_ALL}"
        )
        is_logged_in = self.status_checker.is_logged_in()
        print(
            f"外网访问: {Fore.GREEN if is_logged_in else Fore.RED}{'正常' if is_logged_in else '未登录'}{Style.RESET_ALL}"
        )

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
