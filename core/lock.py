import os
import sys
from pathlib import Path


class FileLock:
    def __init__(self, lock_file: str = "autologin.lock"):
        self.lock_file = Path(lock_file)
        self._locked = False

    def acquire(self) -> bool:
        """获取锁，失败返回False"""
        if self.lock_file.exists():
            try:
                with open(self.lock_file, "r") as f:
                    pid = f.read().strip()
                    if pid:
                        if sys.platform == "win32":
                            import ctypes

                            kernel32 = ctypes.windll.kernel32
                            try:
                                handle = kernel32.OpenProcess(1, False, int(pid))
                                if handle:
                                    kernel32.CloseHandle(handle)
                                    return False
                            except Exception:
                                pass
                        else:
                            try:
                                os.kill(int(pid), 0)
                                return False
                            except OSError:
                                pass
            except Exception:
                pass
        try:
            with open(self.lock_file, "w") as f:
                f.write(str(os.getpid()))
            self._locked = True
            return True
        except Exception:
            return False

    def release(self):
        """释放锁"""
        if self._locked and self.lock_file.exists():
            try:
                self.lock_file.unlink()
            except Exception:
                pass
            self._locked = False

    def is_locked(self) -> bool:
        """检测是否已被锁定"""
        return self.lock_file.exists()

    def __enter__(self):
        if not self.acquire():
            raise RuntimeError("Another instance is already running")
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.release()
