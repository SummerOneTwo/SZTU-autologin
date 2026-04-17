//go:build windows

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf8"
	"unsafe"
)

const taskName = "SZTU-Autologin"

var (
	kernel32AutoStart       = syscall.NewLazyDLL("kernel32.dll")
	procMultiByteToWideChar = kernel32AutoStart.NewProc("MultiByteToWideChar")
)

const (
	codePageGBK           = 936
	seeMaskNoCloseProcess = 0x00000040
	swHide                = 0
)

var elevateAccessDeniedHints = []string{
	"拒绝访问",
	"access is denied",
}

type shellExecuteInfo struct {
	cbSize         uint32
	fMask          uint32
	hwnd           uintptr
	lpVerb         *uint16
	lpFile         *uint16
	lpParameters   *uint16
	lpDirectory    *uint16
	nShow          int32
	hInstApp       uintptr
	lpIDList       uintptr
	lpClass        *uint16
	hkeyClass      uintptr
	dwHotKey       uint32
	hIconOrMonitor uintptr
	hProcess       uintptr
}

var (
	shell32                = syscall.NewLazyDLL("shell32.dll")
	procShellExecuteExW    = shell32.NewProc("ShellExecuteExW")
	procGetExitCodeProcess = kernel32AutoStart.NewProc("GetExitCodeProcess")
	procCloseHandle        = kernel32AutoStart.NewProc("CloseHandle")
)

func enableAutostart() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	// 自启动程序路径
	launchExe := filepath.Join(filepath.Dir(exePath), "sztu-autologin-launch.exe")

	args := []string{
		"/create",
		"/tn", taskName,
		"/tr", fmt.Sprintf(`"%s"`, launchExe),
		"/sc", "onlogon",
		"/f",
	}

	output, err := runSchTasks(args...)
	if err != nil {
		if shouldRetryElevated(output) {
			output, err = runSchTasksElevated(args...)
		}
	}
	if err != nil {
		return fmt.Errorf("创建任务失败: %v\n命令: schtasks %v\n输出: %s", err, args, output)
	}
	return nil
}

func disableAutostart() error {
	args := []string{"/delete", "/tn", taskName, "/f"}

	output, err := runSchTasks(args...)
	if err != nil {
		if shouldRetryElevated(output) {
			output, err = runSchTasksElevated(args...)
		}
	}
	if err != nil {
		return fmt.Errorf("删除任务失败: %v\n命令: schtasks %v\n输出: %s", err, args, output)
	}
	return nil
}

func isAutostartEnabled() bool {
	args := []string{"/query", "/tn", taskName}
	_, err := runSchTasks(args...)
	return err == nil
}

func runSchTasks(args ...string) (string, error) {
	cmd := exec.Command("schtasks", args...)
	output, err := cmd.CombinedOutput()
	return decodeWindowsCommandOutput(output), err
}

func runSchTasksElevated(args ...string) (string, error) {
	tempDir := os.TempDir()
	stdoutPath := filepath.Join(tempDir, "sztu-autologin-schtasks-stdout.txt")
	stderrPath := filepath.Join(tempDir, "sztu-autologin-schtasks-stderr.txt")

	_ = os.Remove(stdoutPath)
	_ = os.Remove(stderrPath)

	commandLine := buildCmdRedirectCommand(args, stdoutPath, stderrPath)

	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString("cmd.exe")
	parameters, _ := syscall.UTF16PtrFromString("/c " + commandLine)

	info := shellExecuteInfo{
		cbSize:       uint32(unsafe.Sizeof(shellExecuteInfo{})),
		fMask:        seeMaskNoCloseProcess,
		lpVerb:       verb,
		lpFile:       file,
		lpParameters: parameters,
		nShow:        swHide,
	}

	ret, _, callErr := procShellExecuteExW.Call(uintptr(unsafe.Pointer(&info)))
	if ret == 0 {
		return "", fmt.Errorf("提权失败: %v", callErr)
	}
	if info.hProcess == 0 {
		return "", fmt.Errorf("提权失败: 未获取到进程句柄")
	}
	defer procCloseHandle.Call(info.hProcess)

	handle := syscall.Handle(info.hProcess)
	_, waitErr := syscall.WaitForSingleObject(handle, syscall.INFINITE)
	if waitErr != nil {
		return "", fmt.Errorf("等待提权进程失败: %v", waitErr)
	}

	var exitCode uint32
	ret, _, callErr = procGetExitCodeProcess.Call(info.hProcess, uintptr(unsafe.Pointer(&exitCode)))
	if ret == 0 {
		return "", fmt.Errorf("获取提权进程退出码失败: %v", callErr)
	}

	output := readElevatedOutput(stdoutPath, stderrPath)
	if exitCode != 0 {
		if output == "" {
			output = fmt.Sprintf("提权后的 schtasks 退出码: %d", exitCode)
		}
		return output, fmt.Errorf("exit status %d", exitCode)
	}

	return output, nil
}

func readElevatedOutput(stdoutPath string, stderrPath string) string {
	parts := make([]string, 0, 2)
	if data, err := os.ReadFile(stdoutPath); err == nil {
		if text := decodeWindowsCommandOutput(data); text != "" {
			parts = append(parts, text)
		}
	}
	if data, err := os.ReadFile(stderrPath); err == nil {
		if text := decodeWindowsCommandOutput(data); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n")
}

func buildCmdRedirectCommand(args []string, stdoutPath string, stderrPath string) string {
	parts := []string{"schtasks"}
	for _, arg := range args {
		parts = append(parts, quoteCmdArg(arg))
	}
	return strings.Join(parts, " ") + " > " + quoteCmdArg(stdoutPath) + " 2> " + quoteCmdArg(stderrPath)
}

func quoteCmdArg(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}

func shouldRetryElevated(output string) bool {
	lowerOutput := strings.ToLower(output)
	for _, hint := range elevateAccessDeniedHints {
		if strings.Contains(lowerOutput, strings.ToLower(hint)) {
			return true
		}
	}
	return false
}

func decodeWindowsCommandOutput(output []byte) string {
	trimmed := bytes.TrimSpace(output)
	if len(trimmed) == 0 {
		return ""
	}

	if utf8.Valid(trimmed) {
		return string(trimmed)
	}

	if decoded, ok := decodeCodePage(trimmed, codePageGBK); ok {
		return decoded
	}

	return string(trimmed)
}

func decodeCodePage(input []byte, codePage uint32) (string, bool) {
	if len(input) == 0 {
		return "", true
	}

	size, _, _ := procMultiByteToWideChar.Call(
		uintptr(codePage),
		0,
		uintptr(unsafe.Pointer(&input[0])),
		uintptr(len(input)),
		0,
		0,
	)
	if size == 0 {
		return "", false
	}

	buf := make([]uint16, size)
	ret, _, _ := procMultiByteToWideChar.Call(
		uintptr(codePage),
		0,
		uintptr(unsafe.Pointer(&input[0])),
		uintptr(len(input)),
		uintptr(unsafe.Pointer(&buf[0])),
		size,
	)
	if ret == 0 {
		return "", false
	}

	return syscall.UTF16ToString(buf), true
}
