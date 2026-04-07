//go:build !windows

package main

import "fmt"

func hideConsoleWindow() {}

func stopDaemon() {
	fmt.Println("stop 命令仅在 Windows 平台可用")
}
