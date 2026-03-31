package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	campusCheckURL   = "http://172.19.0.5"
	internetCheckURL = "http://www.baidu.com"
)

func isOnCampusNetwork() bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(campusCheckURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func isLoggedIn() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(internetCheckURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func runDaemon() {
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

	if !cfg.AutoReconnect {
		fmt.Println("自动重连已禁用，请修改配置启用")
		return
	}

	fmt.Printf("后台守护进程启动，检测间隔: %d秒\n", cfg.CheckInterval)
	fmt.Println("按 Ctrl+C 停止")

	engine := NewLoginEngine(cfg)
	consecutiveFailures := 0
	var pauseUntil time.Time

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(cfg.CheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\n守护进程已停止")
			return
		case <-ticker.C:
			if time.Now().Before(pauseUntil) {
				continue
			}

			if !isOnCampusNetwork() {
				fmt.Printf("[%s] 未连接到校园网\n", time.Now().Format("15:04:05"))
				continue
			}

			if isLoggedIn() {
				consecutiveFailures = 0
				continue
			}

			fmt.Printf("[%s] 检测到断网，尝试重连...\n", time.Now().Format("15:04:05"))
			result := engine.Login()

			if result.Success {
				fmt.Printf("[%s] 重连成功\n", time.Now().Format("15:04:05"))
				consecutiveFailures = 0
			} else {
				fmt.Printf("[%s] 重连失败: %s\n", time.Now().Format("15:04:05"), result.Message)
				consecutiveFailures++

				if consecutiveFailures >= 5 {
					fmt.Printf("[%s] 连续失败5次，暂停10分钟\n", time.Now().Format("15:04:05"))
					pauseUntil = time.Now().Add(10 * time.Minute)
					consecutiveFailures = 0
				}
			}
		}
	}
}
