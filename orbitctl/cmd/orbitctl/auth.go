package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

type apiResponse[T any] struct {
	Success bool   `json:"Success"`
	Data    T      `json:"Data"`
	Message string `json:"Message"`
}

type loginResp struct {
	AccessToken string `json:"access_token"`
	// 可选：若后端返回 refresh_token，也可保存
	RefreshToken string `json:"refresh_token,omitempty"`
}

// CLI认证会话相关结构
type authSessionCreateResp struct {
	SessionID string `json:"session_id"`
	AuthURL   string `json:"auth_url"`
	ExpiresIn int    `json:"expires_in"`
}

type authSessionStatusResp struct {
	Status    string `json:"status"`
	Token     string `json:"token,omitempty"`
	UserEmail string `json:"user_email,omitempty"`
}

func readCredsCLI(user, pass string) (string, string, error) {
	if user != "" && pass != "" {
		return user, pass, nil
	}
	reader := bufio.NewReader(os.Stdin)
	if user == "" {
		fmt.Print("用户名: ")
		u, _ := reader.ReadString('\n')
		user = strings.TrimSpace(u)
	}
	if pass == "" {
		fmt.Print("密码: ")
		if term.IsTerminal(int(syscall.Stdin)) {
			// 在终端中隐藏密码输入
			passBytes, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return "", "", fmt.Errorf("读取密码失败: %w", err)
			}
			pass = string(passBytes)
			fmt.Println() // 输入密码后换行
		} else {
			// 非终端环境(如CI)，使用普通输入
			p, _ := reader.ReadString('\n')
			pass = strings.TrimSpace(p)
		}
	}
	if user == "" || pass == "" {
		return "", "", errors.New("用户名和密码不能为空")
	}
	return user, pass, nil
}

type deviceAuthRequest struct {
	OS          string    `json:"os"`
	DeviceName  string    `json:"device_name"`
	PublicIP    string    `json:"public_ip"`
	RequestTime time.Time `json:"request_time"`
}

func cmdAuthLogin(user, pass, token, apiBase string) error {
	// 如果提供了token，直接保存
	if token != "" {
		if err := saveAccessToken(token); err != nil {
			return fmt.Errorf("保存访问令牌失败: %w", err)
		}
		fmt.Println("已保存访问令牌，登录成功")
		return nil
	}

	// 若通过 --api-base 传入，则覆盖环境变量，集中由 apiURLf 读取
	if apiBase != "" {
		_ = os.Setenv("ORBIT_API_BASE", apiBase)
	}

	fmt.Println("正在创建CLI授权会话...")

	// ================================================================
	// 补全的核心逻辑在这里
	// 1. 获取设备信息
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown" // 提供一个默认值
	}

	if err != nil {
		return fmt.Errorf("获取公网IP失败: %w", err)
	}

	// 2. 构造请求体
	authReq := deviceAuthRequest{
		OS:         runtime.GOOS, // 获取当前操作系统
		DeviceName: hostname,
	}
	// ================================================================

	// 3. 创建授权会话 (修改 httpPostJSON 的第二个参数)
	resp, err := httpPostJSON(apiURL("cli.device_auth.sessions"), authReq, false)
	if err != nil {
		return fmt.Errorf("创建授权会话失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查非 2xx 的状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("创建授权会话失败: 服务器返回状态码 %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	var sessionResp apiResponse[authSessionCreateResp]
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		return fmt.Errorf("解析会话响应失败: %w", err)
	}
	if !sessionResp.Success {
		return fmt.Errorf("创建授权会话失败: %s", sessionResp.Message)
	}

	sessionData := sessionResp.Data
	// fmt.Println(sessionData) // 这一行在实际使用中可以注释掉
	fmt.Printf("✅ 授权会话已创建，会话将在 %d 秒后过期\n", sessionData.ExpiresIn)
	fmt.Printf("请在浏览器中打开以下链接完成登录:\n")
	fmt.Printf("🌐 %s\n\n", sessionData.AuthURL)

	// 尝试自动打开浏览器
	if err := openBrowser(sessionData.AuthURL); err == nil {
		fmt.Println("✅ 已自动在浏览器中打开登录页面")
	} else {
		fmt.Println("💡 无法自动打开浏览器，请手动复制上述链接到浏览器中打开")
	}

	fmt.Println("\n⏳ 等待您在浏览器中完成登录...")

	return pollAuthSession(sessionData.SessionID, sessionData.ExpiresIn)

}

func cmdAuthLogout() error {
	// 尝试调用后端登出API（可选）
	resp, err := httpPostJSON(apiURL("auth.logout"), nil, true)
	if err == nil && resp != nil {
		resp.Body.Close()
	}

	// 清理本地token
	clearTokens()
	fmt.Println("登出成功")
	return nil
}

func cmdAuthRefresh() error {
	if err := refreshAccessToken(); err != nil {
		return fmt.Errorf("刷新令牌失败: %w", err)
	}
	fmt.Println("令牌刷新成功")
	return nil
}

// openBrowser 尝试在默认浏览器中打开URL
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// pollAuthSession 轮询授权会话状态直到成功或超时
// pollAuthSession 轮询认证会话的状态
// @param apiBase   - API服务器的基础URL
// @param sessionID - 当前的会话ID
// @param expiresIn - 会话的有效时间（秒）
func pollAuthSession(sessionID string, expiresIn int) error {
	// 创建一个定时器，每3秒触发一次
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// 创建一个超时channel，在 expiresIn 秒后会接收到信号
	// 这是修复的关键点
	timeout := time.After(time.Duration(expiresIn) * time.Second)

	fmt.Printf("⏳ 正在等待用户授权，请在浏览器中完成操作。此链接将在 %d 秒后过期。\n", expiresIn)
	fmt.Print("进度: ")

	// 使用 for 和 select 监听多个 channel
	for {
		select {
		// 情况一: 超时channel被触发
		case <-timeout:
			fmt.Println() // 换行
			return errors.New("授权会话超时，请重新登录")

		// 情况二: ticker被触发，进行一次轮询检查
		case <-ticker.C:
			resp, err := httpGetJSON(apiURL("cli.device_auth.token", sessionID), false)
			if err != nil { // 辅助函数用于从map中安全获取
				// 打印错误但继续尝试，直到超时
				fmt.Printf("\n⚠️  检查登录状态时出错: %v\n", err)
				fmt.Print("进度: ") // 保持UI一致性
				continue
			}

			var statusResp apiResponse[authSessionStatusResp]
			if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
				resp.Body.Close()
				fmt.Printf("\n⚠️  解析状态响应失败: %v\n", err)
				fmt.Print("进度: ")
				continue // 继续下一次轮询
			}

			if !statusResp.Success {
				resp.Body.Close()
				fmt.Printf("\n⚠️  获取状态失败: %s\n", statusResp.Message)
				fmt.Print("进度: ")
				continue
			}

			status := statusResp.Data
			resp.Body.Close() // 关闭响应体，因为我们已经解析完了

			switch status.Status {
			case "SUCCESS":
				fmt.Println() // 换行
				if err := saveAccessToken(status.Token); err != nil {
					return fmt.Errorf("保存访问令牌失败: %w", err)
				}
				fmt.Printf("✅ 登录成功！用户: %s\n", status.UserEmail)
				return nil // 成功，退出函数
			case "EXPIRED":
				fmt.Println() // 换行
				return errors.New("授权会话已过期，请重新登录")
			case "PENDING":
				fmt.Print(".") // 打印进度点，继续循环
			default:
				fmt.Printf("\n⚠️  未知状态: %s\n", status.Status)
				fmt.Print("进度: ")
			}
		}
	}
}
