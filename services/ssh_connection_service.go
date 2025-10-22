package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/opentdp/go-helper/logman"
	"github.com/opentdp/go-helper/webssh"
	"github.com/youfun/OrbitDeploy/models"
	"golang.org/x/crypto/ssh"
)

// SSHConnectionService SSH连接服务
type SSHConnectionService struct {
	connectionPool map[uuid.UUID]*ssh.Client
	mu             sync.RWMutex
	maxConnections int
	timeout        time.Duration
}

// NewSSHConnectionService 创建SSH连接服务
func NewSSHConnectionService() *SSHConnectionService {
	return &SSHConnectionService{
		connectionPool: make(map[uuid.UUID]*ssh.Client),
		maxConnections: 50, // 最大连接数
		timeout:        30 * time.Second,
	}
}

// GetConnection 获取SSH连接（连接池管理）
func (s *SSHConnectionService) GetConnection(hostID uuid.UUID) (*ssh.Client, error) {
	s.mu.RLock()
	if client, exists := s.connectionPool[hostID]; exists {
		s.mu.RUnlock()
		// 测试连接是否仍然有效
		if s.isConnectionAlive(client) {
			return client, nil
		}
		// 连接已断开，移除并重新创建
		s.removeConnection(hostID)
	} else {
		s.mu.RUnlock()
	}

	// 获取主机信息
	host, err := models.GetSSHHostByID(hostID)
	if err != nil {
		return nil, fmt.Errorf("获取SSH主机信息失败: %w", err)
	}

	// 创建新连接
	client, err := s.createConnection(host)
	if err != nil {
		return nil, fmt.Errorf("创建SSH连接失败: %w", err)
	}

	// 添加到连接池
	s.addConnection(hostID, client)

	return client, nil
}

// TestConnection 测试SSH连接
func (s *SSHConnectionService) TestConnection(hostID uuid.UUID) error {
	host, err := models.GetSSHHostByID(hostID)
	if err != nil {
		return fmt.Errorf("获取SSH主机信息失败: %w", err)
	}

	client, err := s.createConnection(host)
	if err != nil {
		// 更新主机状态为离线
		models.UpdateSSHHostStatus(hostID, "offline")
		return fmt.Errorf("SSH连接失败: %w", err)
	}
	defer client.Close()

	// 测试执行简单命令
	session, err := client.NewSession()
	if err != nil {
		models.UpdateSSHHostStatus(hostID, "error")
		return fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	_, err = session.CombinedOutput("echo 'connection test'")
	if err != nil {
		models.UpdateSSHHostStatus(hostID, "error")
		return fmt.Errorf("SSH命令执行失败: %w", err)
	}

	// 更新主机状态为在线
	models.UpdateSSHHostStatus(hostID, "online")
	logman.Info("SSH连接测试成功", "host_id", hostID)
	return nil
}

// ExecuteCommand 在远程主机上执行命令
func (s *SSHConnectionService) ExecuteCommand(hostID uuid.UUID, command string) (string, error) {
	client, err := s.GetConnection(hostID)
	if err != nil {
		return "", fmt.Errorf("获取SSH连接失败: %w", err)
	}

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建SSH会话失败: %w", err)
	}
	defer session.Close()

	logman.Debug("执行SSH命令", "host_id", hostID, "command", command)

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("命令执行失败: %w", err)
	}

	return string(output), nil
}

// TransferFile 传输文件到远程主机（简化实现）
func (s *SSHConnectionService) TransferFile(hostID uuid.UUID, content, remotePath string) error {
	// 使用echo命令写入文件（生产环境建议使用scp或sftp）
	command := fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", remotePath, content)
	_, err := s.ExecuteCommand(hostID, command)
	if err != nil {
		return fmt.Errorf("文件传输失败: %w", err)
	}

	logman.Info("文件传输成功", "host_id", hostID, "remote_path", remotePath)
	return nil
}

// RefreshHostStatus 刷新主机状态
func (s *SSHConnectionService) RefreshHostStatus(hostID uuid.UUID) error {
	return s.TestConnection(hostID)
}

// RefreshAllHostsStatus 刷新所有主机状态
func (s *SSHConnectionService) RefreshAllHostsStatus() error {
	hosts, err := models.GetActiveSSHHosts()
	if err != nil {
		return fmt.Errorf("获取活跃SSH主机失败: %w", err)
	}

	var wg sync.WaitGroup
	for _, host := range hosts {
		wg.Add(1)
		go func(hostID uuid.UUID) {
			defer wg.Done()
			if err := s.TestConnection(hostID); err != nil {
				logman.Warn("主机状态刷新失败", "host_id", hostID, "error", err)
			}
		}(host.ID)
	}

	wg.Wait()
	return nil
}

// GetConnectionStats 获取连接池统计信息
func (s *SSHConnectionService) GetConnectionStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"active_connections": len(s.connectionPool),
		"max_connections":    s.maxConnections,
		"timeout":            s.timeout.Seconds(),
	}
}

// CloseConnection 关闭指定主机的连接
func (s *SSHConnectionService) CloseConnection(hostID uuid.UUID) {
	s.removeConnection(hostID)
}

// CloseAllConnections 关闭所有连接
func (s *SSHConnectionService) CloseAllConnections() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for hostID, client := range s.connectionPool {
		if client != nil {
			client.Close()
		}
		delete(s.connectionPool, hostID)
	}

	logman.Info("已关闭所有SSH连接")
}

// createConnection 创建SSH连接
func (s *SSHConnectionService) createConnection(host *models.SSHHost) (*ssh.Client, error) {
	// 构建地址
	addr := host.Addr
	if host.Port != 22 {
		addr = fmt.Sprintf("%s:%d", host.Addr, host.Port)
	} else if !containsPort(addr) {
		addr = addr + ":22"
	}

	// 创建SSH客户端选项
	option := &webssh.SSHClientOption{
		Addr:       addr,
		User:       host.User,
		Password:   host.Password,
		PrivateKey: host.PrivateKey,
	}

	// 使用webssh库创建连接
	client, err := webssh.NewSSHClient(option)
	if err != nil {
		return nil, fmt.Errorf("创建SSH客户端失败: %w", err)
	}

	logman.Debug("SSH连接创建成功", "host_id", host.ID, "host_name", host.Name)
	return client, nil
}

// addConnection 添加连接到连接池
func (s *SSHConnectionService) addConnection(hostID uuid.UUID, client *ssh.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查连接池是否已满
	if len(s.connectionPool) >= s.maxConnections {
		// 移除最旧的连接（简化实现，实际应使用LRU）
		for id, oldClient := range s.connectionPool {
			oldClient.Close()
			delete(s.connectionPool, id)
			break
		}
	}

	s.connectionPool[hostID] = client
}

// removeConnection 从连接池移除连接
func (s *SSHConnectionService) removeConnection(hostID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client, exists := s.connectionPool[hostID]; exists {
		if client != nil {
			client.Close()
		}
		delete(s.connectionPool, hostID)
	}
}

// isConnectionAlive 检查连接是否仍然有效
func (s *SSHConnectionService) isConnectionAlive(client *ssh.Client) bool {
	if client == nil {
		return false
	}

	// 尝试创建一个新会话来测试连接
	session, err := client.NewSession()
	if err != nil {
		return false
	}
	defer session.Close()

	// 执行简单的命令测试
	err = session.Run("true")
	return err == nil
}

// containsPort 检查地址是否包含端口
func containsPort(addr string) bool {
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return true
		}
		if addr[i] < '0' || addr[i] > '9' {
			return false
		}
	}
	return false
}

// PeriodicHealthCheck 定期健康检查
func (s *SSHConnectionService) PeriodicHealthCheck(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		logman.Debug("开始定期SSH健康检查")

		// 检查连接池中的连接
		s.mu.RLock()
		hostIDs := make([]uuid.UUID, 0, len(s.connectionPool))
		for hostID := range s.connectionPool {
			hostIDs = append(hostIDs, hostID)
		}
		s.mu.RUnlock()

		// 测试每个连接
		for _, hostID := range hostIDs {
			if err := s.TestConnection(hostID); err != nil {
				logman.Warn("定期健康检查失败，移除连接", "host_id", hostID, "error", err)
				s.removeConnection(hostID)
			}
		}
	}
}

// GetHostResourceInfo 获取主机资源信息
func (s *SSHConnectionService) GetHostResourceInfo(hostID uuid.UUID) (*HostResourceInfo, error) {
	// CPU信息
	cpuInfo, err := s.ExecuteCommand(hostID, "nproc")
	if err != nil {
		return nil, fmt.Errorf("获取CPU信息失败: %w", err)
	}

	// 内存信息（以GB为单位）
	memInfo, err := s.ExecuteCommand(hostID, "free -g | awk 'NR==2{print $2}'")
	if err != nil {
		return nil, fmt.Errorf("获取内存信息失败: %w", err)
	}

	// 磁盘信息（以GB为单位）
	diskInfo, err := s.ExecuteCommand(hostID, "df -BG / | awk 'NR==2{print $2}' | sed 's/G//'")
	if err != nil {
		return nil, fmt.Errorf("获取磁盘信息失败: %w", err)
	}

	// 负载信息
	loadInfo, err := s.ExecuteCommand(hostID, "uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//'")
	if err != nil {
		return nil, fmt.Errorf("获取负载信息失败: %w", err)
	}

	return &HostResourceInfo{
		CPUCores:    cpuInfo,
		MemoryGB:    memInfo,
		DiskGB:      diskInfo,
		LoadAverage: loadInfo,
	}, nil
}

// HostResourceInfo 主机资源信息
type HostResourceInfo struct {
	CPUCores    string `json:"cpu_cores"`
	MemoryGB    string `json:"memory_gb"`
	DiskGB      string `json:"disk_gb"`
	LoadAverage string `json:"load_average"`
}
