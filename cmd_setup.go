package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func readLine(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(password), nil
}

func runSetup() {
	// 设置信号处理，允许用户通过 Ctrl+C 中断
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n操作已取消")
		os.Exit(0)
	}()

	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v，将使用默认配置\n", err)
		cfg = DefaultConfig()
	}
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n欢迎使用 SZTU 校园网自动登录工具！")
	fmt.Println("请按提示完成初始配置: ")

	// Username
	username, err := readLine(reader, "[步骤 1/4] 输入学号: ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	if username == "" {
		fmt.Fprintln(os.Stderr, "错误: 学号不能为空")
		return
	}
	cfg.Username = username

	// Password
	password, err := readPassword("[步骤 2/4] 输入密码: ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取密码失败: %v\n", err)
		return
	}
	if password == "" {
		fmt.Fprintln(os.Stderr, "错误: 密码不能为空")
		return
	}
	cfg.Password = password

	// ISP
	fmt.Println("\n[步骤 3/4] 选择运营商:")
	fmt.Println("  [1] 中国联通 (@cucc)")
	fmt.Println("  [2] 中国移动 (@cmcc)")
	fmt.Println("  [3] 中国电信 (@chinanet)")
	ispChoice, err := readLine(reader, "> 选择 [1]: ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	switch ispChoice {
	case "2":
		cfg.ISP = "cmcc"
	case "3":
		cfg.ISP = "chinanet"
	case "":
		cfg.ISP = "cucc"
	default:
		if ispChoice != "1" {
			fmt.Println("无效选择，使用默认: 中国联通")
		}
		cfg.ISP = "cucc"
	}

	// Area
	fmt.Println("\n[步骤 4/4] 选择区域:")
	fmt.Println("  [1] 宿舍区 (ac_id=17)")
	fmt.Println("  [2] 教学区 (ac_id=1)")
	areaChoice, err := readLine(reader, "> 选择 [1]: ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		return
	}
	switch areaChoice {
	case "2":
		cfg.Area = "teaching"
		cfg.ACID = "1"
	case "":
		cfg.Area = "dormitory"
		cfg.ACID = "17"
	default:
		if areaChoice != "1" {
			fmt.Println("无效选择，使用默认: 宿舍区")
		}
		cfg.Area = "dormitory"
		cfg.ACID = "17"
	}

	// Save config
	if err := SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "保存配置失败: %v\n", err)
		return
	}

	fmt.Println("\n配置已保存，正在测试登录...")

	// Test login
	engine := NewLoginEngine(cfg)
	result := engine.Login()
	if result.Success {
		fmt.Println("✓ 登录成功！")
	} else {
		fmt.Printf("✗ 登录失败: %s\n", result.Message)
	}
}
