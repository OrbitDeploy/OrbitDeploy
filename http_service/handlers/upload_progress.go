package handlers

import (
	"context"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/opentdp/go-helper/logman"
	"github.com/opentdp/go-helper/socket"
	"golang.org/x/net/websocket"
)

// UploadStage represents different stages of file upload
type UploadStage string

const (
	UploadStageStarting  UploadStage = "starting"
	UploadStageUploading UploadStage = "uploading"
	UploadStageSaving    UploadStage = "saving"
	UploadStageCompleted UploadStage = "completed"
	UploadStageFailed    UploadStage = "failed"
)

// UploadProgress represents the current upload progress
type UploadProgress struct {
	UploadId      string      `json:"upload_id"`
	FileName      string      `json:"file_name"`
	Stage         UploadStage `json:"stage"`
	Message       string      `json:"message"`
	Progress      int         `json:"progress"` // 0-100
	BytesUploaded int64       `json:"bytes_uploaded"`
	TotalBytes    int64       `json:"total_bytes"`
	Timestamp     time.Time   `json:"timestamp"`
	Error         string      `json:"error,omitempty"`
}

// UploadSession manages a single upload session
type UploadSession struct {
	UploadId string
	WsConn   *socket.WsConn
	Context  context.Context
	Cancel   context.CancelFunc
	mutex    sync.RWMutex
	progress *UploadProgress
}

// Global upload sessions management
var (
	uploadSessions      = make(map[string]*UploadSession)
	uploadSessionsMutex sync.RWMutex
)

// SendProgress sends the current progress to the WebSocket client
func (us *UploadSession) SendProgress() error {
	us.mutex.RLock()
	progress := *us.progress
	us.mutex.RUnlock()

	data := socket.PlainData{
		Method:  "upload_progress",
		TaskId:  0, // Using upload_id in payload instead
		Success: progress.Stage != UploadStageFailed,
		Message: progress.Message,
		Payload: progress,
	}

	logman.Info("发送上传进度到客户端", "upload_id", us.UploadId, "stage", progress.Stage, "progress", progress.Progress)
	return us.WsConn.WriteJson(data)
}

// UpdateProgress updates the upload progress and sends it to the client
func (us *UploadSession) UpdateProgress(stage UploadStage, message string, progress int, bytesUploaded, totalBytes int64) error {
	us.mutex.Lock()
	us.progress.Stage = stage
	us.progress.Message = message
	us.progress.Progress = progress
	us.progress.BytesUploaded = bytesUploaded
	us.progress.TotalBytes = totalBytes
	us.progress.Timestamp = time.Now()
	us.progress.Error = ""
	us.mutex.Unlock()

	logman.Info("更新上传进度", "upload_id", us.UploadId, "stage", stage, "progress", progress, "message", message)
	return us.SendProgress()
}

// UpdateProgressWithError updates the upload progress with an error
func (us *UploadSession) UpdateProgressWithError(stage UploadStage, message string, err error) error {
	us.mutex.Lock()
	us.progress.Stage = UploadStageFailed
	us.progress.Message = message
	us.progress.Error = err.Error()
	us.progress.Timestamp = time.Now()
	us.mutex.Unlock()

	logman.Error("上传过程中发生错误", "upload_id", us.UploadId, "stage", stage, "message", message, "error", err)
	return us.SendProgress()
}

// GetUploadSession retrieves an upload session by upload ID
func GetUploadSession(uploadId string) *UploadSession {
	uploadSessionsMutex.RLock()
	defer uploadSessionsMutex.RUnlock()
	return uploadSessions[uploadId]
}

// UploadProgressWS handles WebSocket connections for real-time upload progress
func UploadProgressWS(ws *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get upload ID from query parameters
	uploadId := ws.Request().URL.Query().Get("upload_id")
	if uploadId == "" {
		logman.Error("上传进度 WebSocket 缺少 upload_id 参数")
		ws.Close()
		return
	}

	logman.Info("上传进度 WebSocket 已连接", "upload_id", uploadId, "remote", ws.Request().RemoteAddr)

	// Create WsConn wrapper around the websocket connection
	wsConn := &socket.WsConn{Conn: ws}
	defer wsConn.Close()

	// Create upload session
	session := &UploadSession{
		UploadId: uploadId,
		WsConn:   wsConn,
		Context:  ctx,
		Cancel:   cancel,
		progress: &UploadProgress{
			UploadId:  uploadId,
			Stage:     UploadStageStarting,
			Message:   "WebSocket connected, waiting for upload to start...",
			Progress:  0,
			Timestamp: time.Now(),
		},
	}

	// Register session
	uploadSessionsMutex.Lock()
	uploadSessions[uploadId] = session
	logman.Info("已注册上传会话", "upload_id", uploadId)
	uploadSessionsMutex.Unlock()

	// Clean up session on disconnect
	defer func() {
		uploadSessionsMutex.Lock()
		delete(uploadSessions, uploadId)
		logman.Info("已清理上传会话", "upload_id", uploadId)
		uploadSessionsMutex.Unlock()
	}()

	// Send initial progress
	logman.Info("发送初始上传进度", "upload_id", uploadId)
	session.SendProgress()

	// Keep connection alive and handle disconnection
	for {
		select {
		case <-ctx.Done():
			logman.Info("上传 WebSocket 连接已关闭", "upload_id", uploadId)
			return
		default:
			// Check if connection is still alive by trying to read (with timeout)
			ws.SetReadDeadline(time.Now().Add(30 * time.Second))
			var dummy []byte
			_, err := ws.Read(dummy)
			if err != nil {
				logman.Info("上传 WebSocket 客户端断开连接", "upload_id", uploadId)
				return
			}
			time.Sleep(5 * time.Second)
		}
	}
}

// UploadProgressHandler serves the WebSocket endpoint for upload progress
func UploadProgressHandler(c echo.Context) error {
	logman.Info("收到上传进度 WebSocket 请求")
	websocket.Handler(UploadProgressWS).ServeHTTP(c.Response().Writer, c.Request())
	return nil
}

// Helper function to send upload progress for an upload ID
func SendUploadProgress(uploadId string, stage UploadStage, message string, progress int, bytesUploaded, totalBytes int64) {
	logman.Info("通过上传ID发送上传进度", "upload_id", uploadId, "stage", stage, "progress", progress)
	session := GetUploadSession(uploadId)
	if session != nil {
		session.UpdateProgress(stage, message, progress, bytesUploaded, totalBytes)
	}
}

// Helper function to send upload error for an upload ID
func SendUploadError(uploadId string, message string, err error) {
	logman.Error("通过上传ID发送上传错误", "upload_id", uploadId, "message", message, "error", err)
	session := GetUploadSession(uploadId)
	if session != nil {
		session.UpdateProgressWithError(UploadStageFailed, message, err)
	}
}
