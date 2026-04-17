// 自启动启动器 - 开机时静默启动守护进程
// 编译为 GUI 程序，不显示任何窗口
package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

type Config struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	ISP           string `json:"isp"`
	ACID          string `json:"ac_id"`
	AutoReconnect bool   `json:"auto_reconnect"`
	CheckInterval int    `json:"check_interval"`
}

func main() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}

	dir := filepath.Dir(exePath)
	mainExe := filepath.Join(dir, "sztu-autologin.exe")
	configPath := filepath.Join(dir, "config.json")

	// 读取并更新配置
	configData, err := os.ReadFile(configPath)
	if err == nil {
		var cfg Config
		if json.Unmarshal(configData, &cfg) == nil {
			cfg.AutoReconnect = true
			if newData, err := json.MarshalIndent(cfg, "", "  "); err == nil {
				os.WriteFile(configPath, newData, 0644)
			}
		}
	}

	// 启动守护进程
	cmd := exec.Command(mainExe, "daemon", "-hide")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x08000000,
	}
	cmd.Start()
}
