package handlers

import (
	"log"
	"net/http"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"github.com/OrbitDeploy/OrbitDeploy/services"
	"github.com/labstack/echo/v4"
)

// Routing Handlers

func CreateRoutingHandler(c echo.Context) error {
	appIDStr := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
	if err != nil {
		log.Printf("解析应用ID失败: %v", err)
		return SendError(c, http.StatusBadRequest, "Invalid application ID format")
	}

	var req RoutingRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("绑定请求体失败: %v", err)
		return SendError(c, http.StatusBadRequest, "Invalid request body")
	}

	// 先调用域名处理服务进行冲突检查和 Caddy 配置
	message, cleanDomain, httpErr := services.ManageRouting(appID, req.DomainName, req.HostPort, "add")
	if httpErr != nil {
		log.Printf("域名处理服务失败，应用ID: %s, 域名: %s, 错误: %v", appID, req.DomainName, httpErr)
		return SendError(c, httpErr.Code, httpErr.Message.(string))
	}
	log.Printf("域名处理服务成功: %s", message)

	// 所有检查和配置都成功后，再创建数据库记录
	routing, err := models.CreateRouting(appID, cleanDomain, req.HostPort, req.IsActive)
	if err != nil {
		log.Printf("创建路由记录失败，应用ID: %s, 域名: %s, 错误: %v", appID, cleanDomain, err)
		// 如果数据库创建失败，需要回滚 Caddy 配置
		rollbackMessage, _, rollbackErr := services.ManageRouting(appID, cleanDomain, req.HostPort, "remove")
		if rollbackErr != nil {
			log.Printf("回滚 Caddy 配置失败: %v", rollbackErr)
		} else {
			log.Printf("回滚成功: %s", rollbackMessage)
		}
		return SendError(c, http.StatusInternalServerError, "Failed to create routing record")
	}

	return SendCreated(c, map[string]interface{}{
		"routing":     toRoutingResponse(routing),
		"cleanDomain": cleanDomain,
		"message":     message,
	})
}

func ListRoutingsByAppHandler(c echo.Context) error {
	appIDStr := c.Param("appId")
	appID, err := DecodeFriendlyID(PrefixApplication, appIDStr)
	if err != nil {
		log.Printf("解析应用ID失败: %v", err)
		return SendError(c, http.StatusBadRequest, "Invalid application ID format")
	}

	routings, err := models.ListRoutingsByAppID(appID)
	if err != nil {
		log.Printf("获取路由列表失败，应用ID: %s, 错误: %v", appID, err)
		return SendError(c, http.StatusInternalServerError, "Failed to list routings")
	}

	// Convert to RoutingResponse slice for response
	var routingResponses []*RoutingResponse
	for _, r := range routings {
		routingResponses = append(routingResponses, toRoutingResponse(r))
	}

	log.Printf("成功获取应用路由列表，应用ID: %s, 路由数量: %d", appID, len(routings))
	return SendSuccess(c, map[string]interface{}{
		"routings": routingResponses,
		"message":  "Routings retrieved successfully",
	})
}

func UpdateRoutingHandler(c echo.Context) error {
	routingIDStr := c.Param("routingId")
	routingID, err := DecodeFriendlyID(PrefixRouting, routingIDStr)
	if err != nil {
		log.Printf("解析路由ID失败: %v", err)
		return SendError(c, http.StatusBadRequest, "Invalid routing ID format")
	}

	var req RoutingRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("绑定更新请求体失败: %v", err)
		return SendError(c, http.StatusBadRequest, "Invalid request body")
	}

	// 调用服务层更新路由
	routing, err := services.UpdateRouting(routingID, req.DomainName, req.HostPort, req.IsActive)
	if err != nil {
		log.Printf("更新路由失败，路由ID: %s, 域名: %s, 错误: %v", routingID, req.DomainName, err)
		return SendError(c, http.StatusInternalServerError, err.Error())
	}

	return SendSuccess(c, map[string]interface{}{
		"routing": toRoutingResponse(routing),
		"message": "Routing updated successfully",
	})
}

func DeleteRoutingHandler(c echo.Context) error {
	routingIDStr := c.Param("routingId")
	routingID, err := DecodeFriendlyID(PrefixRouting, routingIDStr)
	if err != nil {
		log.Printf("解析路由ID失败: %v", err)
		return SendError(c, http.StatusBadRequest, "Invalid routing ID format")
	}

	// 调用服务层删除路由
	err = services.DeleteRouting(routingID)
	if err != nil {
		log.Printf("删除路由失败，路由ID: %s, 错误: %v", routingID, err)
		return SendError(c, http.StatusInternalServerError, err.Error())
	}

	return SendSuccess(c, map[string]interface{}{
		"message": "Routing deleted successfully",
	})
}

func toRoutingResponse(r *models.Routing) *RoutingResponse {
	return &RoutingResponse{
		Uid:            EncodeFriendlyID(PrefixRouting, r.ID),
		ApplicationUid: EncodeFriendlyID(PrefixApplication, r.ApplicationID),
		DomainName:     r.DomainName,
		HostPort:       r.HostPort,
		IsActive:       r.IsActive,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}
