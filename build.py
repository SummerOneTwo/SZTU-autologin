#!/usr/bin/env python3
import subprocess
from pathlib import Path


def build():
    print("安装依赖...")
    subprocess.run(["uv", "add", "requests", "colorama"], check=True)

    print("\n使用 PyInstaller 打包...")
    subprocess.run(
        ["uv", "run", "pyinstaller", "-F", "-w", "--name", "SZTU-Autologin", "main.py"],
        check=True,
    )

    dist_path = Path("dist")
    if dist_path.exists():
        print(f"\n打包完成！可执行文件位于: {dist_path.absolute()}")
        exe_file = dist_path / "SZTU-Autologin.exe"
        if exe_file.exists():
            print(f"可执行文件: {exe_file}")


if __name__ == "__main__":
    build()
