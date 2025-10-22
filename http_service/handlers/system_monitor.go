package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/opentdp/go-helper/logman"
	"github.com/opentdp/go-helper/psutil"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/tmaxmax/go-sse"
)

// DiskPartition represents a disk partition with usage information
type DiskPartition struct {
	Device     string `json:"device"`
	Mountpoint string `json:"mountpoint"`
	Fstype     string `json:"fstype"`
	Total      uint64 `json:"total"`
	Used       uint64 `json:"used"`
}

// ExtendedSystemStats extends the standard psutil summary with disk information
type ExtendedSystemStats struct {
	*psutil.SummaryStat
	DiskPartitions []DiskPartition `json:"disk_partitions"`
	DiskTotal      uint64          `json:"disk_total"`
	DiskUsed       uint64          `json:"disk_used"`
}

// SSE server and publisher for system monitoring
var systemSSEServer *sse.Server
var systemSSEOnce sync.Once
var systemPublisherOnce sync.Once

func getSystemSSEServer() *sse.Server {
	systemSSEOnce.Do(func() {
		// Create a Joe provider; replayer is not required for live stats
		joe := &sse.Joe{}
		systemSSEServer = &sse.Server{
			Provider: joe,
			OnSession: func(w http.ResponseWriter, r *http.Request) (topics []string, allowed bool) {
				logman.Info("System monitor SSE connected", "remote", r.RemoteAddr)
				// All system monitor clients subscribe to a common topic
				return []string{"system_monitor"}, true
			},
		}
		logman.Info("System monitor SSE server initialized")
	})
	return systemSSEServer
}

// startSystemStatsPublisher ensures a single publisher goroutine sends updates periodically
func startSystemStatsPublisher() {
	systemPublisherOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				stats := getExtendedSystemStats()
				payload, _ := json.Marshal(stats)
				msg := &sse.Message{Type: sse.Type("system-stats")}
				msg.AppendData(string(payload))
				// Publish to the common topic
				server := getSystemSSEServer()
				if err := server.Publish(msg, "system_monitor"); err != nil {
					logman.Warn("Failed to publish system stats via SSE", "error", err)
				}
				<-ticker.C
			}
		}()
	})
}

// getExtendedSystemStats gets system statistics including disk partition information
func getExtendedSystemStats() *ExtendedSystemStats {
	// Get basic system summary
	basicStats := psutil.Summary(false) // false means don't fetch public IP addresses for performance

	// Get disk partitions
	dp, err := disk.Partitions(false)
	if err != nil {
		logman.Error("Failed to get disk partitions", "error", err)
		// Return basic stats without disk info if disk enumeration fails
		return &ExtendedSystemStats{
			SummaryStat:    basicStats,
			DiskPartitions: []DiskPartition{},
			DiskTotal:      0,
			DiskUsed:       0,
		}
	}

	// Collect disk partition information
	diskPartition := []DiskPartition{}
	diskTotaled := ","
	diskTotal := uint64(0)
	diskUsed := uint64(0)

	for _, dpi := range dp {
		du, err := disk.Usage(dpi.Mountpoint)
		if err != nil {
			logman.Warn("Failed to get disk usage", "mountpoint", dpi.Mountpoint, "error", err)
			continue
		}

		if du.Total > 0 || du.Used > 0 {
			diskPartition = append(diskPartition, DiskPartition{
				Device:     dpi.Device,
				Mountpoint: dpi.Mountpoint,
				Fstype:     dpi.Fstype,
				Total:      du.Total,
				Used:       du.Used,
			})
		}

		if !strings.Contains(diskTotaled, dpi.Device) {
			diskTotaled += dpi.Device + ","
			diskTotal += du.Total
			diskUsed += du.Used
		}
	}

	return &ExtendedSystemStats{
		SummaryStat:    basicStats,
		DiskPartitions: diskPartition,
		DiskTotal:      diskTotal,
		DiskUsed:       diskUsed,
	}
}

// SystemMonitorHandler serves the SSE endpoint for system monitoring
func SystemMonitorHandler(c echo.Context) error {
	// Start the publisher if not already started
	startSystemStatsPublisher()
	// Serve the SSE stream
	server := getSystemSSEServer()
	server.ServeHTTP(c.Response().Writer, c.Request())
	return nil
}
