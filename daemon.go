package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	campusCheckURL   = "http://172.19.0.5"
	internetCheckURL = "http://www.baidu.com"
)

var (
	DaemonClient = &http.Client{Timeout: 5 * time.Second}
	// 不自动跟随重定向的客户端，用于检测是否被劫持
	NoRedirectClient = &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
)

func isOnCampusNetwork() bool {
	resp, err := DaemonClient.Get(campusCheckURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode == 200
}

// isLoggedIn 检查是否已登录校园网
// 未认证时访问外网会被重定向到登录页面，需要检测这种情况
func isLoggedIn() bool {
	resp, err := NoRedirectClient.Get(internetCheckURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == 302 || resp.StatusCode == 301 {
		location := resp.Header.Get("Location")
		if strings.Contains(location, "172.19.0.5") || strings.Contains(location, "srun") {
			return false
		}
	}

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

	fmt.Println("=======================================")
	fmt.Println("    SZTU 校园网自动登录 - 守护进程")
	fmt.Println("=======================================")
	fmt.Printf("用户: %s (%s)\n", cfg.Username, getISPName(cfg.ISP))
	fmt.Printf("检测间隔: %d 秒\n", cfg.CheckInterval)
	fmt.Println("---------------------------------------")
	fmt.Println("按 Ctrl+C 停止")
	fmt.Println()

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
