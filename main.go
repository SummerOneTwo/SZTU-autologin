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
	case "autostart":
		runAutostart()
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
	fmt.Println("  sztu-autologin setup              交互式配置")
	fmt.Println("  sztu-autologin login              立即登录")
	fmt.Println("  sztu-autologin daemon             后台运行（自动重连）")
	fmt.Println("  sztu-autologin autostart [on|off|status]  开机自启动管理")
	fmt.Println("  sztu-autologin help               显示帮助")
}

func runLogin() {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "配置无效: %v\n", err)
		fmt.Fprintln(os.Stderr, "请先运行: sztu-autologin setup")
		os.Exit(1)
	}

	fmt.Println("正在登录...")
	engine := NewLoginEngine(cfg)
	result := engine.Login()

	if result.Success {
		fmt.Printf("✓ %s\n", result.Message)
	} else {
		fmt.Printf("✗ %s\n", result.Message)
		os.Exit(1)
	}
}

func runAutostart() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: sztu-autologin autostart [on|off|status]")
		return
	}
	switch os.Args[2] {
	case "on":
		if err := enableAutostart(); err != nil {
			fmt.Fprintf(os.Stderr, "启用失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("开机自启动已启用")
	case "off":
		if err := disableAutostart(); err != nil {
			fmt.Fprintf(os.Stderr, "禁用失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("开机自启动已禁用")
	case "status":
		if isAutostartEnabled() {
			fmt.Println("开机自启动: 已启用")
		} else {
			fmt.Println("开机自启动: 未启用")
		}
	default:
		fmt.Println("Usage: sztu-autologin autostart [on|off|status]")
	}
}
