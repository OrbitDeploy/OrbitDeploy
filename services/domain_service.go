package services

import (
	"fmt"
	"log"
	"strings"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/OrbitDeploy/OrbitDeploy/utils"
	"github.com/OrbitDeploy/fastcaddy"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ManageRouting 管理应用路由配置（服务层）
func ManageRouting(applicationID uuid.UUID, domain string, port int, action string) (message string, cleanDomain string, httpErr *echo.HTTPError) {
	var err error

	// 验证并清理域名（去除协议前缀）
	cleanDomain, err = utils.NormalizeDomain(domain)
	if err != nil {
		fmt.Printf("❌ [路由操作] 域名格式无效: %v\n", err)
		return "", "", echo.NewHTTPError(400, fmt.Sprintf("无效的域名格式: %v", err))
	}
	fmt.Printf("✅ [路由操作] 域名清理完成: %s -> %s\n", domain, cleanDomain)

	// 初始化 FastCaddy 客户端
	fc := fastcaddy.New()

	switch action {
	case "add":
		if port == 0 {
			return "", "", echo.NewHTTPError(400, "添加路由时端口是必需的")
		}

		// 对于 xxx.xxx.com 格式的域名，强制使用 8080 端口
		// if isDomainWithStandardFormat(cleanDomain) {
		// 	port = 8080
		// }

		proxyTo := fmt.Sprintf("localhost:%d", port) // 假设代理到本地端口

		fmt.Printf("🔧 [路由添加] 开始处理域名: %s -> %s\n", cleanDomain, proxyTo)

		// 检查域名冲突（全局范围）
		fmt.Printf("🔍 [路由添加] 检查域名冲突: %s\n", cleanDomain)
		if exists, err := checkDomainConflict(cleanDomain); err != nil {
			fmt.Printf("❌ [路由添加] 域名冲突检查失败: %v\n", err)
			return "", "", echo.NewHTTPError(500, fmt.Sprintf("检查域名冲突时出错: %v", err))
		} else if exists {
			fmt.Printf("❌ [路由添加] 域名冲突: %s\n", cleanDomain)
			return "", "", echo.NewHTTPError(400, "域名已存在")
		}
		fmt.Printf("✅ [路由添加] 域名冲突检查通过\n")

		// 检查端口冲突（应用级别）
		fmt.Printf("🔍 [路由添加] 检查端口冲突: %d (应用ID: %s)\n", port, applicationID)
		if exists, err := checkPortConflict(port, applicationID); err != nil {
			fmt.Printf("❌ [路由添加] 端口冲突检查失败: %v\n", err)
			return "", "", echo.NewHTTPError(500, fmt.Sprintf("检查端口冲突时出错: %v", err))
		} else if exists {
			fmt.Printf("❌ [路由添加] 端口冲突: %d (应用ID: %s)\n", port, applicationID)
			return "", "", echo.NewHTTPError(400, "该应用下端口已存在")
		}
		fmt.Printf("✅ [路由添加] 端口冲突检查通过\n")

		// 使用 FastCaddy 添加路由
		fmt.Printf("🚀 [路由添加] 通过 FastCaddy 添加路由配置: %s -> %s\n", cleanDomain, proxyTo)
		err = fc.AddReverseProxy(cleanDomain, proxyTo)
		if err != nil {
			fmt.Printf("❌ [路由添加] FastCaddy 添加失败: %v\n", err)
			return "", "", echo.NewHTTPError(500, fmt.Sprintf("通过 Caddy 添加路由失败: %v", err))
		}

		message = fmt.Sprintf("路由 %s 配置成功", cleanDomain)
		fmt.Printf("🎉 [路由添加] 完成: %s\n", message)

	case "remove":
		fmt.Printf("🗑️ [路由删除] 开始处理路由删除: %s\n", cleanDomain)

		// 使用 FastCaddy 删除路由
		fmt.Printf("🚀 [路由删除] 通过 FastCaddy 删除路由: %s\n", cleanDomain)
		err = fc.DeleteRoute(cleanDomain)
		if err != nil {
			fmt.Printf("❌ [路由删除] FastCaddy 删除失败: %v\n", err)
			return "", "", echo.NewHTTPError(500, fmt.Sprintf("通过 Caddy 删除路由失败: %v", err))
		}

		// 从数据库删除路由记录
		fmt.Printf("💾 [路由删除] 从数据库删除路由记录\n")
		routings, err := models.ListRoutings()
		if err != nil {
			log.Printf("获取路由配置失败: %v", err)
		} else {
			for _, routing := range routings {
				if routing.DomainName == cleanDomain && routing.ApplicationID == applicationID {
					err := models.DeleteRouting(routing.ID)
					if err != nil {
						fmt.Printf("❌ [路由删除] 数据库删除失败: %v\n", err)
						log.Printf("删除路由记录失败: %v", err)
					} else {
						fmt.Printf("✅ [路由删除] 数据库删除成功\n")
					}
					break
				}
			}
		}

		message = fmt.Sprintf("路由 %s 删除成功", cleanDomain)
		fmt.Printf("🎉 [路由删除] 完成: %s\n", message)

	default:
		return "", "", echo.NewHTTPError(400, "无效的操作。使用 'add' 或 'remove'")
	}

	return message, cleanDomain, nil
}

// UpdateRouting 更新路由配置（服务层）
func UpdateRouting(routingID uuid.UUID, newDomain string, newPort int, isActive bool) (*models.Routing, error) {
	// 获取旧的路由记录
	oldRouting, err := models.GetRoutingByID(routingID)
	if err != nil {
		return nil, fmt.Errorf("获取路由记录失败: %v", err)
	}

	// 验证并清理新域名
	cleanDomain, err := utils.NormalizeDomain(newDomain)
	if err != nil {
		return nil, fmt.Errorf("域名格式无效: %v", err)
	}

	// 初始化 FastCaddy 客户端
	fc := fastcaddy.New()

	// 删除旧的 Caddy 配置
	fmt.Printf("🚀 [路由更新] 删除旧的 Caddy 配置: %s\n", oldRouting.DomainName)
	err = fc.DeleteRoute(oldRouting.DomainName)
	if err != nil {
		return nil, fmt.Errorf("删除旧的 Caddy 配置失败: %v", err)
	}

	// 检查新域名的冲突（如果域名改变）
	if cleanDomain != oldRouting.DomainName {
		fmt.Printf("🔍 [路由更新] 检查新域名冲突: %s\n", cleanDomain)
		if exists, err := checkDomainConflict(cleanDomain); err != nil {
			// 回滚：添加回旧的配置
			_ = fc.AddReverseProxy(oldRouting.DomainName, fmt.Sprintf("localhost:%d", oldRouting.HostPort))
			return nil, fmt.Errorf("检查域名冲突失败: %v", err)
		} else if exists {
			// 回滚
			_ = fc.AddReverseProxy(oldRouting.DomainName, fmt.Sprintf("localhost:%d", oldRouting.HostPort))
			return nil, fmt.Errorf("新域名已存在: %s", cleanDomain)
		}
	}

	// 检查新端口的冲突（如果端口改变）
	if newPort != oldRouting.HostPort {
		fmt.Printf("🔍 [路由更新] 检查新端口冲突: %d (应用ID: %s)\n", newPort, oldRouting.ApplicationID)
		if exists, err := checkPortConflict(newPort, oldRouting.ApplicationID); err != nil {
			// 回滚
			_ = fc.AddReverseProxy(oldRouting.DomainName, fmt.Sprintf("localhost:%d", oldRouting.HostPort))
			return nil, fmt.Errorf("检查端口冲突失败: %v", err)
		} else if exists {
			// 回滚
			_ = fc.AddReverseProxy(oldRouting.DomainName, fmt.Sprintf("localhost:%d", oldRouting.HostPort))
			return nil, fmt.Errorf("该应用下新端口已存在: %d", newPort)
		}
	}

	// 添加新的 Caddy 配置
	proxyTo := fmt.Sprintf("localhost:%d", newPort)
	fmt.Printf("🚀 [路由更新] 添加新的 Caddy 配置: %s -> %s\n", cleanDomain, proxyTo)
	err = fc.AddReverseProxy(cleanDomain, proxyTo)
	if err != nil {
		// 回滚：添加回旧的配置
		_ = fc.AddReverseProxy(oldRouting.DomainName, fmt.Sprintf("localhost:%d", oldRouting.HostPort))
		return nil, fmt.Errorf("添加新的 Caddy 配置失败: %v", err)
	}

	// 更新数据库
	fmt.Printf("💾 [路由更新] 更新数据库\n")
	return models.UpdateRouting(routingID, cleanDomain, newPort, isActive)
}

// DeleteRouting 删除路由配置（服务层）
func DeleteRouting(routingID uuid.UUID) error {
	// 获取路由记录
	routing, err := models.GetRoutingByID(routingID)
	if err != nil {
		return fmt.Errorf("获取路由记录失败: %v", err)
	}

	// 初始化 FastCaddy 客户端
	fc := fastcaddy.New()

	// 删除 Caddy 配置
	fmt.Printf("🚀 [路由删除] 删除 Caddy 配置: %s\n", routing.DomainName)
	err = fc.DeleteRoute(routing.DomainName)
	if err != nil {
		return fmt.Errorf("删除 Caddy 配置失败: %v", err)
	}

	// 删除数据库记录
	fmt.Printf("💾 [路由删除] 删除数据库记录\n")
	return models.DeleteRouting(routingID)
}

// checkDomainConflict 检查域名是否已存在（全局范围）
func checkDomainConflict(domain string) (bool, error) {
	routings, err := models.ListRoutings()
	if err != nil {
		return false, err
	}
	for _, routing := range routings {
		if routing.DomainName == domain {
			return true, nil
		}
	}
	return false, nil
}

// checkPortConflict 检查端口是否已存在（应用级别）
func checkPortConflict(port int, applicationID uuid.UUID) (bool, error) {
	routings, err := models.ListRoutings()
	if err != nil {
		return false, err
	}
	for _, routing := range routings {
		if routing.HostPort == port && routing.ApplicationID == applicationID {
			return true, nil
		}
	}
	return false, nil
}

// isDomainWithStandardFormat 检查是否为标准域名格式（例如 xxx.xxx.com）
func isDomainWithStandardFormat(domain string) bool {
	// 简单检查：包含至少两个点，且不以点开头或结尾
	parts := strings.Split(domain, ".")
	return len(parts) >= 3 && parts[0] != "" && parts[len(parts)-1] != ""
}
