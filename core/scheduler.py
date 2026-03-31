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
            "schtasks",
            "/create",
            "/tn",
            cls.TASK_NAME,
            "/tr",
            f"{exe_path} --silent",
            "/sc",
            "onlogon",
            "/rl",
            "highest",
            "/f",
        ]

        try:
            result = subprocess.run(cmd, capture_output=True, text=True)
            return result.returncode == 0
        except Exception:
            return False

    @classmethod
    def delete_task(cls) -> bool:
        """删除任务"""
        cmd = ["schtasks", "/delete", "/tn", cls.TASK_NAME, "/f"]
        try:
            result = subprocess.run(cmd, capture_output=True, text=True)
            return result.returncode == 0
        except Exception:
            return False

    @classmethod
    def task_exists(cls) -> bool:
        """检查任务是否存在"""
        cmd = ["schtasks", "/query", "/tn", cls.TASK_NAME]
        try:
            result = subprocess.run(cmd, capture_output=True, text=True)
            return result.returncode == 0
        except Exception:
            return False
