//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
)

const taskName = "SZTU-Autologin"

func enableAutostart() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command("schtasks",
		"/create",
		"/tn", taskName,
		"/tr", fmt.Sprintf(`"%s" daemon`, exePath),
		"/sc", "onlogon",
		"/rl", "highest",
		"/f",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("创建任务失败: %v, 输出: %s", err, output)
	}
	return nil
}

func disableAutostart() error {
	cmd := exec.Command("schtasks", "/delete", "/tn", taskName, "/f")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("删除任务失败: %v, 输出: %s", err, output)
	}
	return nil
}

func isAutostartEnabled() bool {
	cmd := exec.Command("schtasks", "/query", "/tn", taskName)
	return cmd.Run() == nil
}
