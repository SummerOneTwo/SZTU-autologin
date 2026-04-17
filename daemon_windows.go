//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var (
	user32DLL        = syscall.NewLazyDLL("user32.dll")
	procShowWindow   = user32DLL.NewProc("ShowWindow")
	procGetConsoleWindow = user32DLL.NewProc("GetConsoleWindow")
)

const SW_HIDE = 0

// Windows 进程创建标志
const CREATE_NO_WINDOW = 0x08000000

func hideConsoleWindow() {
	defer func() {
		recover()
	}()
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd != 0 {
		procShowWindow.Call(hwnd, SW_HIDE)
	}
}

func isDaemonRunning() bool {
	exePath, err := os.Executable()
	if err != nil {
		return false
	}
	exeName := strings.ToLower(filepath.Base(exePath))
	currentPid := os.Getpid()

	// 使用 PowerShell 获取进程命令行，以区分 daemon 和其他子命令
	psScript := fmt.Sprintf("Get-WmiObject Win32_Process -Filter \"name='%s'\" | Select-Object ProcessId, CommandLine | ConvertTo-Csv -NoTypeInformation", exeName)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	output, err := cmd.Output()
	if err != nil {
		return isDaemonRunningSimple()
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "\"ProcessId\"") {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) < 2 {
			continue
		}
		pid := strings.Trim(fields[0], `"`)
		cmdLine := strings.Trim(fields[1], `"`)
		if pid == "" || pid == fmt.Sprintf("%d", currentPid) {
			continue
		}
		if strings.Contains(cmdLine, " daemon") {
			return true
		}
	}
	return false
}

func isDaemonRunningSimple() bool {
	exePath, err := os.Executable()
	if err != nil {
		return false
	}
	exeName := strings.ToLower(filepath.Base(exePath))
	currentPid := os.Getpid()

	cmd := exec.Command("tasklist", "/fo", "csv", "/nh")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) < 2 {
			continue
		}
		name := strings.Trim(fields[0], `"`)
		pid := strings.Trim(fields[1], `"`)
		if strings.ToLower(name) == exeName && pid != fmt.Sprintf("%d", currentPid) {
			return true
		}
	}
	return false
}

func startDaemonHidden() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %w", err)
	}

	if isDaemonRunning() {
		return fmt.Errorf("守护进程已在运行")
	}

	cmd := exec.Command(exePath, "daemon", "-hide")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | CREATE_NO_WINDOW,
	}
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动进程失败: %w", err)
	}

	time.Sleep(2 * time.Second)

	if !isDaemonRunning() {
		return fmt.Errorf("守护进程启动后立即退出")
	}

	return nil
}

func stopDaemon() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取程序路径失败: %v\n", err)
		os.Exit(1)
	}
	exeName := strings.ToLower(filepath.Base(exePath))

	cmd := exec.Command("tasklist", "/fo", "csv", "/nh")
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取进程列表失败: %v\n", err)
		os.Exit(1)
	}

	var pids []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) < 2 {
			continue
		}
		name := strings.Trim(fields[0], `"`)
		pid := strings.Trim(fields[1], `"`)
		if strings.ToLower(name) == exeName && pid != fmt.Sprintf("%d", os.Getpid()) {
			pids = append(pids, pid)
		}
	}

	if len(pids) == 0 {
		fmt.Println("守护进程未运行")
		return
	}

	for _, pid := range pids {
		cmd := exec.Command("taskkill", "/pid", pid, "/f")
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "终止进程 %s 失败: %v\n", pid, err)
		} else {
			fmt.Printf("已终止进程 %s\n", pid)
		}
	}
}
