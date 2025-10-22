package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/dborm"
	"github.com/youfun/OrbitDeploy/models"
)

// DockerBuildQueueService Docker构建队列服务，用于防止多个镜像同时触发构建
type DockerBuildQueueService struct {
	maxConcurrent   int
	pollingInterval time.Duration
	buildService    *BuildService
}

// NewDockerBuildQueueService 创建新的Docker构建队列服务
func NewDockerBuildQueueService(maxConcurrent int, pollingInterval time.Duration) *DockerBuildQueueService {
	return &DockerBuildQueueService{
		maxConcurrent:   maxConcurrent,
		pollingInterval: pollingInterval,
		buildService:    NewBuildService(),
	}
}

// InitDB 初始化数据库表
func (dbqs *DockerBuildQueueService) InitDB() error {
	// 自动迁移表
	if err := dborm.Db.AutoMigrate(&models.DockerBuildTask{}); err != nil {
		return fmt.Errorf("迁移Docker构建任务表失败: %w", err)
	}
	return nil
}

// RecoverTasks 恢复中断的任务
func (dbqs *DockerBuildQueueService) RecoverTasks() error {
	log.Println("正在检查是否有中断的Docker构建任务需要恢复...")
	res := dborm.Db.Model(&models.DockerBuildTask{}).Where("status = ?", models.DockerBuildStatusRunning).Update("status", models.DockerBuildStatusPending)
	if res.Error != nil {
		return fmt.Errorf("恢复任务时出错: %w", res.Error)
	}
	recoveredCount := res.RowsAffected
	if recoveredCount > 0 {
		log.Printf("已成功恢复 %d 个中断的Docker构建任务。\n", recoveredCount)
	} else {
		log.Println("没有需要恢复的Docker构建任务。")
	}
	return nil
}

// StartWorkers 启动工作协程
func (dbqs *DockerBuildQueueService) StartWorkers(ctx context.Context, wg *sync.WaitGroup) {
	log.Printf("启动 %d 个Docker构建工作协程...\n", dbqs.maxConcurrent)
	for i := 1; i <= dbqs.maxConcurrent; i++ {
		wg.Add(1)
		go dbqs.worker(ctx, wg, i)
	}
}

// SubmitDockerBuildTask 提交Docker构建任务
func (dbqs *DockerBuildQueueService) SubmitDockerBuildTask(appID, releaseID uuid.UUID, logText, dockerfile, contextPath string, buildArgs map[string]string) (string, error) {
	payload := models.DockerBuildPayload{
		AppID:       appID,
		ReleaseID:   releaseID,
		LogText:     logText,
		Dockerfile:  dockerfile,
		ContextPath: contextPath,
		BuildArgs:   buildArgs,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("任务序列化失败: %w", err)
	}
	taskID := uuid.New().String()
	_, err = models.CreateDockerBuildTask(taskID, string(payloadBytes), models.DockerBuildStatusPending)
	if err != nil {
		log.Printf("Docker构建任务入队失败: %v", err)
		return "", fmt.Errorf("任务入队失败: %w", err)
	}
	log.Printf("新Docker构建任务已入队: %s (AppID: %s, ReleaseID: %s)\n", taskID, appID, releaseID)
	return taskID, nil
}

// worker 工作协程
func (dbqs *DockerBuildQueueService) worker(ctx context.Context, wg *sync.WaitGroup, id int) {
	defer wg.Done()
	log.Printf("Docker构建工作协程 %d 已启动\n", id)
	for {
		select {
		case <-ctx.Done():
			log.Printf("Docker构建工作协程 %d 收到停机信号，即将退出。\n", id)
			return
		default:
			task, err := models.DequeueDockerBuildTask()
			if err != nil {
				log.Printf("Docker构建工作协程 %d: 出队操作失败: %v\n", id, err)
				time.Sleep(dbqs.pollingInterval)
				continue
			}
			if task == nil {
				time.Sleep(dbqs.pollingInterval)
				continue
			}
			log.Printf("Docker构建工作协程 %d: 开始处理任务 %s\n", id, task.UUID)
			err = dbqs.processDockerBuildTask(task)
			if err != nil {
				log.Printf("Docker构建工作协程 %d: 任务 %s 处理失败: %v\n", id, task.UUID, err)
				models.UpdateDockerBuildTaskStatus(task.UUID, models.DockerBuildStatusFailed, err.Error())
			} else {
				log.Printf("Docker构建工作协程 %d: 任务 %s 处理成功!\n", id, task.UUID)
				models.UpdateDockerBuildTaskStatus(task.UUID, models.DockerBuildStatusCompleted, "")
				// 生成构建日志
				dbqs.generateBuildLog(task)
			}
		}
	}
}

// processDockerBuildTask 处理Docker构建任务
func (dbqs *DockerBuildQueueService) processDockerBuildTask(task *models.DockerBuildTask) error {
	var payload models.DockerBuildPayload
	if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
		return fmt.Errorf("反序列化任务失败: %w", err)
	}

	// 获取应用和release信息
	app, err := models.GetApplicationByID(payload.AppID)
	if err != nil {
		return fmt.Errorf("获取应用失败: %w", err)
	}
	release, err := models.GetReleaseByID(payload.ReleaseID)
	if err != nil {
		return fmt.Errorf("获取release失败: %w", err)
	}

	// 调用实际的构建逻辑
	log.Printf("开始构建Docker镜像 %s (Release: %s)\n", app.Name, release.ImageName)

	buildReq := BuildFromApplicationRequest{
		ApplicationID: payload.AppID,
		Dockerfile:    payload.Dockerfile,
		ContextPath:   payload.ContextPath,
		BuildArgs:     payload.BuildArgs,
	}

	imageName, err := dbqs.buildService.BuildImageFromApplication(buildReq)
	if err != nil {
		return fmt.Errorf("docker构建失败: %w", err)
	}

	log.Printf("Docker镜像构建完成 %s, 镜像名称: %s\n", app.Name, imageName)
	return nil
}

// generateBuildLog 生成构建日志
func (dbqs *DockerBuildQueueService) generateBuildLog(task *models.DockerBuildTask) {
	var payload models.DockerBuildPayload
	if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
		log.Printf("生成构建日志失败: 反序列化错误 %v", err)
		return
	}

	app, err := models.GetApplicationByID(payload.AppID)
	if err != nil {
		log.Printf("生成构建日志失败: 获取应用错误 %v", err)
		return
	}
	release, err := models.GetReleaseByID(payload.ReleaseID)
	if err != nil {
		log.Printf("生成构建日志失败: 获取release错误 %v", err)
		return
	}

	// 生成中文构建日志
	logContent := fmt.Sprintf(`# Docker构建日志

## 构建信息
- **任务ID**: %s
- **应用名称**: %s
- **应用ID**: %s
- **Release ID**: %s
- **镜像名称**: %s
- **构建时间**: %s
- **状态**: 成功

## 构建参数
- **Dockerfile**: %s
- **上下文路径**: %s
- **构建参数**: %v

## 构建内容
- 从GitHub仓库构建了新的Docker镜像版本
- 更新了应用代码和依赖

## 构建日志
%s

---
*此日志由系统自动生成*
`, task.UUID, app.Name, payload.AppID, payload.ReleaseID, release.ImageName, time.Now().Format("2006-01-02 15:04:05"), payload.Dockerfile, payload.ContextPath, payload.BuildArgs, payload.LogText)

	// 写入文件
	fileName := fmt.Sprintf("docker_build_log_%s_%d.md", task.UUID, time.Now().Unix())
	filePath := fmt.Sprintf("build_logs/%s", fileName)
	if err := os.MkdirAll("build_logs", 0755); err != nil {
		log.Printf("创建build_logs目录失败: %v", err)
		return
	}
	if err := os.WriteFile(filePath, []byte(logContent), 0644); err != nil {
		log.Printf("写入构建日志失败: %v", err)
		return
	}
	log.Printf("构建日志已生成: %s", filePath)
}
