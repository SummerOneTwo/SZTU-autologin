package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func runSetup() {
	cfg, _ := LoadConfig()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n欢迎使用 SZTU 校园网自动登录工具！")
	fmt.Println("请按提示完成初始配置：\n")

	// Username
	fmt.Print("[步骤 1/4] 输入学号: ")
	username, _ := reader.ReadString('\n')
	cfg.Username = strings.TrimSpace(username)

	// Password
	fmt.Print("[步骤 2/4] 输入密码: ")
	password, _ := reader.ReadString('\n')
	cfg.Password = strings.TrimSpace(password)

	// ISP
	fmt.Println("\n[步骤 3/4] 选择运营商:")
	fmt.Println("  [1] 中国联通 (@cucc)")
	fmt.Println("  [2] 中国移动 (@cmcc)")
	fmt.Println("  [3] 中国电信 (@chinanet)")
	fmt.Print("> 选择: ")
	ispChoice, _ := reader.ReadString('\n')
	switch strings.TrimSpace(ispChoice) {
	case "2":
		cfg.ISP = "cmcc"
	case "3":
		cfg.ISP = "chinanet"
	default:
		cfg.ISP = "cucc"
	}

	// Area
	fmt.Println("\n[步骤 4/4] 选择区域:")
	fmt.Println("  [1] 宿舍区 (ac_id=17)")
	fmt.Println("  [2] 教学区 (ac_id=1)")
	fmt.Print("> 选择: ")
	areaChoice, _ := reader.ReadString('\n')
	if strings.TrimSpace(areaChoice) == "2" {
		cfg.Area = "teaching"
		cfg.ACID = "1"
	} else {
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
