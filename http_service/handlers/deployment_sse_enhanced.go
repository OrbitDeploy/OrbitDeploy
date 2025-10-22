package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/opentdp/go-helper/logman"
	"github.com/tmaxmax/go-sse"
	"github.com/youfun/OrbitDeploy/models"
)

// 部署日志消息结构体
type DeploymentLogMessage struct {
	DeploymentUid string `json:"deployment_uid"`
	Message       string `json:"message"`
	Timestamp     string `json:"timestamp"`
	Status        string `json:"status,omitempty"`
	Completed     bool   `json:"completed,omitempty"`
	Final         bool   `json:"final,omitempty"`
}

// 全局SSE服务器实例，为每个部署ID提供独立的topic
var deploymentSSEServer *sse.Server
var sseServerOnce sync.Once

// 获取全局SSE服务器
func getDeploymentSSEServer() *sse.Server {
	sseServerOnce.Do(func() {
		// 创建FiniteReplayer，最多回放50条历史消息
		replayer, err := sse.NewFiniteReplayer(50, true) // autoIDs=true，自动生成消息ID
		if err != nil {
			logman.Error("创建FiniteReplayer失败", "error", err)
			replayer = nil
		}

		// 创建Joe提供者，配置replayer
		joe := &sse.Joe{
			Replayer: replayer,
		}

		// 创建SSE服务器
		deploymentSSEServer = &sse.Server{
			Provider: joe,
			OnSession: func(w http.ResponseWriter, r *http.Request) (topics []string, allowed bool) {
				// 从URL路径中提取deployment_id
				pathParts := strings.Split(r.URL.Path, "/")
				var deploymentIDStr string
				for i, part := range pathParts {
					if part == "deployments" && i+1 < len(pathParts) {
						deploymentIDStr = pathParts[i+1]
						break
					}
				}

				if deploymentIDStr == "" {
					logman.Error("SSE会话中缺少deployment_id")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("deployment_id is required"))
					return nil, false
				}

				deploymentID, err := DecodeFriendlyID(PrefixDeployment, deploymentIDStr)
				if err != nil {
					logman.Error("SSE会话中deployment_id格式无效", "error", err)
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("Invalid deployment_id format"))
					return nil, false
				}

				// 验证部署是否存在
				deployment, err := models.GetDeploymentByID(deploymentID)
				if err != nil {
					logman.Error("SSE会话中部署不存在", "deployment_id", deploymentID, "error", err)
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte("Deployment not found"))
					return nil, false
				}

				logman.Info("新的部署日志SSE会话已授权", "deployment_id", deploymentID, "remote", r.RemoteAddr)

				// 如果部署已完成，将历史日志添加到Replayer
				if (deployment.Status == "success" || deployment.Status == "failed") && deployment.LogText != "" {
					go func() {
						// 异步添加历史消息到Replayer
						addCompletedDeploymentLogsToReplayer(deploymentID, deployment, joe)
					}()
				}

				// 订阅到该部署的topic
				topic := fmt.Sprintf("deployment_%s", deploymentID.String())
				return []string{topic}, true
			},
		}

		logman.Info("部署SSE服务器已初始化")
	})
	return deploymentSSEServer
}

// 将已完成部署的历史日志添加到Replayer
func addCompletedDeploymentLogsToReplayer(deploymentID uuid.UUID, deployment *models.Deployment, joe *sse.Joe) {
	if deployment.LogText == "" {
		return
	}

	topic := fmt.Sprintf("deployment_%s", deploymentID.String())

	// 解析并添加已保存的日志到Replayer
	logLines := strings.Split(deployment.LogText, "\n")
	for _, line := range logLines {
		if strings.TrimSpace(line) != "" {
			logMsg := DeploymentLogMessage{
				DeploymentUid: EncodeFriendlyID(PrefixDeployment, deploymentID),
				Message:      line,
				Timestamp:    time.Now().Format("15:04:05"),
				Completed:    true,
			}
			jsonData, _ := json.Marshal(logMsg)

			// 创建SSE消息并添加到Replayer
			sseMessage := &sse.Message{
				Type: sse.Type("message"),
			}
			sseMessage.AppendData(string(jsonData))

			// 通过Provider发布消息（这会自动添加到Replayer）
			joe.Publish(sseMessage, []string{topic})
		}
	}

	// 发送完成状态
	completionMsg := DeploymentLogMessage{
		DeploymentUid: EncodeFriendlyID(PrefixDeployment, deploymentID),
		Status:       deployment.Status,
		Completed:    true,
		Final:        true,
		Timestamp:    time.Now().Format("15:04:05"),
	}
	jsonData, _ := json.Marshal(completionMsg)

	sseMessage := &sse.Message{
		Type: sse.Type("message"),
	}
	sseMessage.AppendData(string(jsonData))
	joe.Publish(sseMessage, []string{topic})

	logman.Info("已完成部署的历史日志已添加到Replayer", "deployment_id", deploymentID, "status", deployment.Status)
}

// 增强版部署日志SSE处理器，支持历史消息回放
// GET /api/deployments/{deploymentId}/logs
func DeploymentLogsSSEEnhanced(c echo.Context) error {
	// Support both parameter formats: :deploymentId and :deployment_id
	deploymentIDStr := c.Param("deploymentId")
	if deploymentIDStr == "" {
		deploymentIDStr = c.Param("deployment_id")
	}
	if deploymentIDStr == "" {
		return echo.NewHTTPError(400, "deploymentId is required")
	}

	// 验证deploymentId格式
	deploymentID, err := DecodeFriendlyID(PrefixDeployment, deploymentIDStr)
	if err != nil {
		return echo.NewHTTPError(400, "Invalid deploymentId format")
	}

	// Validate application token permission if using app token
	authType := c.Get("auth_type")
	if authType == "app_token" {
		// Get deployment to check application ownership
		deployment, err := models.GetDeploymentByID(deploymentID)
		if err != nil {
			return echo.NewHTTPError(404, "Deployment not found")
		}

		if err := validateApplicationTokenPermission(c, deployment.ApplicationID); err != nil {
			return err
		}
	}

	// 获取SSE服务器并处理请求
	server := getDeploymentSSEServer()
	server.ServeHTTP(c.Response().Writer, c.Request())
	return nil
}

// 发送部署日志消息到SSE客户端
func SendDeploymentLogSSE(deploymentID uuid.UUID, message string) {
	server := getDeploymentSSEServer()

	logMsg := DeploymentLogMessage{
		DeploymentUid: EncodeFriendlyID(PrefixDeployment, deploymentID),
		Message:      message,
		Timestamp:    time.Now().Format("15:04:05"),
	}

	jsonData, _ := json.Marshal(logMsg)

	// 创建SSE消息
	sseMessage := &sse.Message{
		Type: sse.Type("message"),
	}
	sseMessage.AppendData(string(jsonData))

	// 发送到该部署的topic
	topic := fmt.Sprintf("deployment_%s", deploymentID.String())
	if err := server.Publish(sseMessage, topic); err != nil {
		logman.Error("发送部署日志SSE失败", "deployment_id", deploymentID, "error", err)
	} else {
		logman.Info("发送部署日志SSE", "deployment_id", deploymentID, "message", message)
	}
}

// 发送部署状态更新到SSE客户端
func SendDeploymentStatusSSE(deploymentID uuid.UUID, status string, final bool) {
	server := getDeploymentSSEServer()

	statusMsg := DeploymentLogMessage{
		DeploymentUid: EncodeFriendlyID(PrefixDeployment, deploymentID),
		Status:       status,
		Timestamp:    time.Now().Format("15:04:05"),
		Final:        final,
	}

	jsonData, _ := json.Marshal(statusMsg)

	// 创建SSE消息
	sseMessage := &sse.Message{
		Type: sse.Type("message"),
	}
	sseMessage.AppendData(string(jsonData))

	// 发送到该部署的topic
	topic := fmt.Sprintf("deployment_%s", deploymentID.String())
	if err := server.Publish(sseMessage, topic); err != nil {
		logman.Error("发送部署状态SSE失败", "deployment_id", deploymentID, "error", err)
	} else {
		logman.Info("发送部署状态SSE", "deployment_id", deploymentID, "status", status, "final", final)
	}
}

// 保存部署日志并发送最终状态（替代原有的SSE保存函数）
func SaveDeploymentLogsSSEEnhanced(deploymentID uuid.UUID, status string, logText string) {
	// 更新数据库
	models.UpdateDeployment(deploymentID, status, logText, &time.Time{})

	// 发送最终状态
	SendDeploymentStatusSSE(deploymentID, status, true)

	logman.Info("保存部署日志到数据库（增强版SSE）", "deployment_id", deploymentID, "status", status)
}