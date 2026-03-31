package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "setup":
		runSetup()
	case "login":
		runLogin()
	case "daemon":
		runDaemon()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("SZTU 校园网自动登录工具")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  sztu-autologin setup    交互式配置")
	fmt.Println("  sztu-autologin login    立即登录")
	fmt.Println("  sztu-autologin daemon   后台运行（自动重连）")
	fmt.Println("  sztu-autologin help     显示帮助")
}

// Placeholder functions - will be implemented in later tasks
func runLogin() {
	fmt.Println("Login command - to be implemented")
}

func runDaemon() {
	fmt.Println("Daemon command - to be implemented")
}
