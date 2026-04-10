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
	user32DLL            = syscall.NewLazyDLL("user32.dll")
	procGetConsoleWindow = user32DLL.NewProc("GetConsoleWindow")
	procShowWindow       = user32DLL.NewProc("ShowWindow")
)

const SW_HIDE = 0

// Windows 进程创建标志
const DETACHED_PROCESS = 0x00000008

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

	// 先检查当前是否有守护进程在运行
	if isDaemonRunning() {
		return fmt.Errorf("守护进程已在运行")
	}

	// 使用 cmd /c start 来启动完全独立的进程
	// 这样父进程退出时子进程不会被终止
	cmd := exec.Command("cmd", "/c", "start", "/b", "", exePath, "daemon", "-hide")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动进程失败: %w", err)
	}

	// 等待进程启动
	time.Sleep(2 * time.Second)

	// 检查进程是否仍在运行
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

	// 查找同名进程
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
		// CSV 格式: "ImageName","PID","SessionName","Session#","MemUsage"
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

	// 终止找到的进程
	for _, pid := range pids {
		cmd := exec.Command("taskkill", "/pid", pid, "/f")
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "终止进程 %s 失败: %v\n", pid, err)
		} else {
			fmt.Printf("已终止进程 %s\n", pid)
		}
	}
}
