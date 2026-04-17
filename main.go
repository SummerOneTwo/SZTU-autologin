package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type cliArgs struct {
	command string
	action  string
	hide    bool
}

type menuStatus struct {
	onCampus         bool
	loggedIn         bool
	autostartEnabled bool
}

func main() {
	args := parseCLIArgs(os.Args[1:])

	// GUI 模式下，交互模式需要附加父控制台
	if args.command == "" {
		attachParentConsole()
		runInteractive()
		return
	}

	// daemon 模式仍需隐藏窗口（作为双重保障）
	if args.hide && args.command == "daemon" {
		hideConsoleWindow()
	}

	switch args.command {
	case "setup":
		runSetup()
	case "login":
		runLogin()
	case "daemon":
		runDaemon()
	case "stop":
		stopDaemon()
	case "autostart":
		runAutostart(args.action)
	case "autostart-launch":
		runAutostartLaunch()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args.command)
		printUsage()
		os.Exit(1)
	}
}

func parseCLIArgs(rawArgs []string) cliArgs {
	parsed := cliArgs{}
	nonFlagArgs := make([]string, 0, len(rawArgs))

	for _, arg := range rawArgs {
		if arg == "-hide" {
			parsed.hide = true
		} else {
			nonFlagArgs = append(nonFlagArgs, arg)
		}
	}

	if len(nonFlagArgs) > 0 {
		parsed.command = nonFlagArgs[0]
	}
	if parsed.command == "autostart" && len(nonFlagArgs) > 1 {
		parsed.action = nonFlagArgs[1]
	}

	return parsed
}

func runInteractive() {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		waitExit()
		return
	}

	// 检查配置与进程状态一致性
	daemonRunning := isDaemonRunning()
	if cfg.AutoReconnect && !daemonRunning {
		// 配置为 true 但进程不存在 → 修正配置
		cfg.AutoReconnect = false
		if err := SaveConfig(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "修正配置失败: %v\n", err)
		}
	} else if !cfg.AutoReconnect && daemonRunning {
		// 配置为 false 但守护进程仍在运行 → 停止守护进程
		stopDaemon()
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "配置无效: %v\n", err)
		fmt.Fprintln(os.Stderr, "请先选择 [1] 修改账号密码")
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		clearScreen()

		status := collectMenuStatus()
		showStatus(cfg, status)
		showMenu(cfg, status)
		fmt.Print("输入选项: ")
		choice, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
			continue
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			runSetup()
			cfg, err = LoadConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			}
		case "2":
			doLogin(cfg)
		case "3":
			toggleAutoReconnect(&cfg)
		case "4":
			toggleAutostart()
		case "5":
			// 刷新状态，直接进入下一次循环
			continue
		case "0":
			fmt.Println("再见！")
			return
		default:
			fmt.Println("无效选项")
		}

		if choice != "0" && choice != "5" {
			fmt.Print("\n按回车继续...")
			_, err = reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
			}
		}
	}
}

func collectMenuStatus() menuStatus {
	return menuStatus{
		onCampus:         isOnCampusNetwork(),
		loggedIn:         isLoggedIn(),
		autostartEnabled: isAutostartEnabled(),
	}
}

func showStatus(cfg Config, status menuStatus) {
	fmt.Println("\n=======================================")
	fmt.Println("    SZTU 校园网自动登录工具")
	fmt.Println("=======================================")

	autostartStatus := "未启用"
	if status.autostartEnabled {
		autostartStatus = "已启用"
	}

	fmt.Printf("网络状态: %s\n", getNetworkStatus(status))
	fmt.Printf("开机自启: %s\n", autostartStatus)

	if cfg.Username == "" {
		fmt.Println("配置状态: 未配置")
	} else {
		fmt.Printf("账号: %s (%s)\n", cfg.Username, getISPName(cfg.ISP))
		fmt.Printf("自动重连: %s\n", boolToStatus(cfg.AutoReconnect))
	}

	fmt.Println("---------------------------------------")
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func getNetworkStatus(status menuStatus) string {
	if status.onCampus {
		if status.loggedIn {
			return "已连接 (校园网已登录)"
		}
		return "未登录 (已连接校园网)"
	}
	if status.loggedIn {
		return "已连接 (外网)"
	}
	return "未连接"
}

func boolToStatus(b bool) string {
	if b {
		return "已启用"
	}
	return "已禁用"
}

func showMenu(cfg Config, status menuStatus) {
	fmt.Println("\n[1] 修改账号密码")

	if status.onCampus && status.loggedIn {
		fmt.Println("[2] 重新登录")
	} else {
		fmt.Println("[2] 立即登录")
	}

	if cfg.AutoReconnect {
		fmt.Println("[3] 关闭自动重连")
	} else {
		fmt.Println("[3] 开启自动重连")
	}

	if status.autostartEnabled {
		fmt.Println("[4] 关闭开机自启")
	} else {
		fmt.Println("[4] 开启开机自启")
	}

	fmt.Println("[5] 刷新状态")
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

// ensureDaemonRunning 确保守护进程在运行
// 返回 (已运行, 错误)，用于调用方决定输出消息
func ensureDaemonRunning() (alreadyRunning bool, err error) {
	if isDaemonRunning() {
		return true, nil
	}
	return false, startDaemonHidden()
}

func toggleAutoReconnect(cfg *Config) {
	newState := !cfg.AutoReconnect

	if newState {
		// 开启：保存配置并确保守护进程运行
		cfg.AutoReconnect = true
		if err := SaveConfig(*cfg); err != nil {
			fmt.Fprintf(os.Stderr, "保存失败: %v\n", err)
			return
		}

		alreadyRunning, err := ensureDaemonRunning()
		if err != nil {
			fmt.Fprintf(os.Stderr, "启动守护进程失败: %v\n", err)
			// 回滚配置
			cfg.AutoReconnect = false
			SaveConfig(*cfg)
			return
		}
		if alreadyRunning {
			fmt.Println("自动重连已启用（守护进程已在运行）")
		} else {
			fmt.Println("自动重连已启用（后台运行中）")
		}
	} else {
		// 关闭：保存配置 + 停止守护进程
		cfg.AutoReconnect = false
		if err := SaveConfig(*cfg); err != nil {
			fmt.Fprintf(os.Stderr, "保存失败: %v\n", err)
			return
		}
		stopDaemon()
		fmt.Println("自动重连已关闭")
	}
}

func toggleAutostart() {
	enabled := isAutostartEnabled()
	if enabled {
		if err := disableAutostart(); err != nil {
			fmt.Fprintf(os.Stderr, "关闭失败: %v\n", err)
			return
		}
		fmt.Println("开机自启已关闭")
	} else {
		if err := enableAutostart(); err != nil {
			fmt.Fprintf(os.Stderr, "开启失败: %v\n", err)
			return
		}
		fmt.Println("开机自启已开启")
	}
}

func doLogin(cfg Config) {
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "配置无效: %v\n", err)
		fmt.Fprintln(os.Stderr, "请先选择 [1] 修改账号密码")
		return
	}

	fmt.Printf("正在登录 [%s - %s]...\n", cfg.Username, getISPName(cfg.ISP))
	engine := NewLoginEngine(cfg)
	result := engine.Login()

	if result.Success {
		fmt.Printf("✓ %s\n", result.Message)
	} else {
		fmt.Printf("✗ %s\n", result.Message)
	}
}

func handleAutostartAction(action string) int {
	switch action {
	case "on":
		if err := enableAutostart(); err != nil {
			fmt.Fprintf(os.Stderr, "启用失败: %v\n", err)
			return 1
		}
		fmt.Println("开机自启动已启用")
	case "off":
		if err := disableAutostart(); err != nil {
			fmt.Fprintf(os.Stderr, "禁用失败: %v\n", err)
			return 1
		}
		fmt.Println("开机自启动已禁用")
	case "status":
		if isAutostartEnabled() {
			fmt.Println("开机自启动: 已启用")
		} else {
			fmt.Println("开机自启动: 未启用")
		}
	}
	return 0
}

func waitExit() {
	fmt.Print("\n按回车键退出...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

func printUsage() {
	fmt.Println("SZTU 校园网自动登录工具")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  sztu-autologin              交互式菜单")
	fmt.Println("  sztu-autologin setup        交互式配置")
	fmt.Println("  sztu-autologin login        立即登录")
	fmt.Println("  sztu-autologin daemon       后台运行（自动重连）")
	fmt.Println("  sztu-autologin daemon -hide 后台静默运行（隐藏窗口）")
	fmt.Println("  sztu-autologin stop         停止守护进程")
	fmt.Println("  sztu-autologin autostart [on|off|status]  开机自启动管理")
	fmt.Println("  sztu-autologin help         显示帮助")
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

	fmt.Printf("正在登录 [%s - %s]...\n", cfg.Username, getISPName(cfg.ISP))
	engine := NewLoginEngine(cfg)
	result := engine.Login()

	if result.Success {
		fmt.Printf("✓ %s\n", result.Message)
	} else {
		fmt.Printf("✗ %s\n", result.Message)
		os.Exit(1)
	}
}

func runAutostartLaunch() {
	// GUI 模式下无需隐藏窗口

	// 开机自启动时，先验证配置
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		return
	}

	// 验证配置有效性
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "配置无效: %v\n", err)
		return
	}

	// 写入配置
	cfg.AutoReconnect = true
	if err := SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "保存配置失败: %v\n", err)
		return
	}

	// 确保守护进程运行
	_, _ = ensureDaemonRunning()
}

func runAutostart(action string) {
	if action == "" {
		fmt.Println("Usage: sztu-autologin autostart [on|off|status]")
		return
	}
	if action != "on" && action != "off" && action != "status" {
		fmt.Println("Usage: sztu-autologin autostart [on|off|status]")
		return
	}
	os.Exit(handleAutostartAction(action))
}
