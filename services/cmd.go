package services

import (
	"fmt"

	"github.com/opentdp/go-helper/command"
	"github.com/opentdp/go-helper/logman"
)

// reloadSystemdDaemon reloads the systemd daemon
func ReloadSystemdDaemon() error {
	logman.Info("执行 systemctl daemon-reload 命令")
	_, err := command.Exec(&command.ExecPayload{
		Content:     "systemctl daemon-reload",
		CommandType: "SHELL",
		Timeout:     30,
	})
	if err != nil {
		logman.Error("systemctl daemon-reload 执行失败", "error", err)
		return fmt.Errorf("failed to reload systemd daemon: %w", err)
	}

	logman.Info("systemctl daemon-reload 执行成功")
	return nil
}
