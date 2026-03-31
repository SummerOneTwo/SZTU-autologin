package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		runInteractive()
		return
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

func runInteractive() {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		waitExit()
		return
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		showStatus(cfg)
		showMenu()
		fmt.Print("输入选项: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			runSetup()
			cfg, _ = LoadConfig()
		case "2":
			runLogin()
		case "3":
			toggleAutoReconnect(&cfg)
		case "4":
			runAutostartInteractive()
		case "0":
			fmt.Println("再见！")
			return
		default:
			fmt.Println("无效选项")
		}

		if choice != "0" {
			fmt.Print("\n按回车继续...")
			reader.ReadString('\n')
		}
	}
}

func showStatus(cfg Config) {
	fmt.Println("\n=======================================")
	fmt.Println("    SZTU 校园网自动登录工具 v3.0.0")
	fmt.Println("=======================================")
	fmt.Printf("用户名: %s\n", cfg.Username)
	fmt.Printf("运营商: %s\n", getISPName(cfg.ISP))
	fmt.Printf("区域: %s\n", getAreaName(cfg.Area))
	fmt.Printf("自动重连: %v\n", cfg.AutoReconnect)
	fmt.Println("---------------------------------------")
}

func showMenu() {
	fmt.Println("\n[1] 修改账号密码")
	fmt.Println("[2] 立即登录")
	fmt.Println("[3] 开关自动重连")
	fmt.Println("[4] 开机自启动管理")
	fmt.Println("[0] 退出")
}

func getISPName(isp string) string {
	names := map[string]string{
		"cucc":     "中国联通",
		"cmcc":     "中国移动",
		"chinanet": "中国电信",
	}
	if name, ok := names[isp]; ok {
		return name
	}
	return isp
}

func getAreaName(area string) string {
	names := map[string]string{
		"dormitory": "宿舍区",
		"teaching":  "教学区",
	}
	if name, ok := names[area]; ok {
		return name
	}
	return area
}

func toggleAutoReconnect(cfg *Config) {
	cfg.AutoReconnect = !cfg.AutoReconnect
	if err := SaveConfig(*cfg); err != nil {
		fmt.Fprintf(os.Stderr, "保存失败: %v\n", err)
		return
	}
	status := "已启用"
	if !cfg.AutoReconnect {
		status = "已禁用"
	}
	fmt.Printf("自动重连 %s\n", status)
}

func runAutostartInteractive() {
	fmt.Println("\n开机自启动管理:")
	fmt.Println("  [1] 启用")
	fmt.Println("  [2] 禁用")
	fmt.Println("  [3] 查看状态")
	fmt.Print("> 选择: ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		if err := enableAutostart(); err != nil {
			fmt.Fprintf(os.Stderr, "启用失败: %v\n", err)
		} else {
			fmt.Println("开机自启动已启用")
		}
	case "2":
		if err := disableAutostart(); err != nil {
			fmt.Fprintf(os.Stderr, "禁用失败: %v\n", err)
		} else {
			fmt.Println("开机自启动已禁用")
		}
	case "3":
		if isAutostartEnabled() {
			fmt.Println("开机自启动: 已启用")
		} else {
			fmt.Println("开机自启动: 未启用")
		}
	default:
		fmt.Println("无效选项")
	}
}

func waitExit() {
	fmt.Print("\n按回车键退出...")
	bufio.NewReader(os.Stdin).ReadString('\n')
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
