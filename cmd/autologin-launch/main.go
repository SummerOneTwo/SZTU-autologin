// 自启动启动器 - 开机时静默启动守护进程
// 编译为 GUI 程序（go build -ldflags "-H=windowsgui"），不显示任何窗口。
// 仅作为无窗口转发器，真正的配置与启动逻辑统一在主程序的 autostart-launch 命令中。
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}

	mainExe := filepath.Join(filepath.Dir(exePath), "sztu-autologin.exe")
	cmd := exec.Command(mainExe, "autostart-launch")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x08000000,
	}
	cmd.Start()
}
