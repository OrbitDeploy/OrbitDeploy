import { createSignal, createEffect, onCleanup } from 'solid-js'
import { buildApiUrl, API_ENDPOINTS } from '../api/config'
import { DiskPartition, SystemStats, ConnectionStatus } from '../types/system'

export function useSystemStats() {
  const [stats, setStats] = createSignal<SystemStats | null>(null)
  const [connection, setConnection] = createSignal<ConnectionStatus>({ status: 'connecting' })
  
  let es: EventSource | null = null

  // Format bytes to human readable format
  const formatBytes = (bytes: number): string => {
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    if (bytes === 0) return '0 B'
    const i = Math.floor(Math.log(bytes) / Math.log(1024))
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i]
  }

  // Format uptime to human readable format
  const formatUptime = (seconds: number): string => {
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    
    if (days > 0) {
      return `${days}d ${hours}h ${minutes}m`
    } else if (hours > 0) {
      return `${hours}h ${minutes}m`
    } else {
      return `${minutes}m`
    }
  }

  // Get memory usage percentage
  const getMemoryUsagePercent = (): number => {
    const currentStats = stats()
    if (!currentStats || currentStats.MemoryTotal === 0) return 0
    return Math.round((currentStats.MemoryUsed / currentStats.MemoryTotal) * 100)
  }

  // Get CPU usage percentage
  const getCpuUsagePercent = (): number => {
    const currentStats = stats()
    if (!currentStats || !currentStats.CpuPercent || currentStats.CpuPercent.length === 0) return 0
    return Math.round(currentStats.CpuPercent[0])
  }

  // Get disk usage percentage
  const getDiskUsagePercent = (): number => {
    const currentStats = stats()
    if (!currentStats || !currentStats.disk_total || currentStats.disk_total === 0) return 0
    return Math.round((currentStats.disk_used! / currentStats.disk_total!) * 100)
  }

  // Connect to SSE
  const connectWebSocket = () => {
    try {
      setConnection({ status: 'connecting' })

      // In development, connect directly to backend. In production, use proxy.
      const params = new URLSearchParams()
      try {
        const token = (window as any).getAccessToken?.()
        if (token) params.set('access_token', token)
      } catch {}
      const sseUrl = `${buildApiUrl(API_ENDPOINTS.system.monitor)}${params.toString() ? '?' + params.toString() : ''}`

      es = new EventSource(sseUrl)

      es.onopen = () => {
        console.log('SSE connected')
        setConnection({ status: 'connected', lastUpdate: Date.now() })
      }

      es.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as any
          // The server sends direct SystemStats payloads
          setStats(data as SystemStats)
          setConnection({ status: 'connected', lastUpdate: Date.now() })
        } catch (err) {
          console.error('Failed to parse SSE message:', err)
        }
      }

      es.addEventListener('system-stats', (event: MessageEvent) => {
        try {
          const data = JSON.parse((event as MessageEvent).data) as any
          setStats(data as SystemStats)
          setConnection({ status: 'connected', lastUpdate: Date.now() })
        } catch (err) {
          console.error('Failed to parse system-stats SSE message:', err)
        }
      })

      es.onerror = (error) => {
        console.error('SSE error:', error)
        // The EventSource will automatically attempt reconnection
        setConnection({ status: 'reconnecting' })
      }
    } catch (err) {
      console.error('Failed to connect SSE:', err)
      setConnection({ status: 'disconnected' })
    }
  }

  // Cleanup on component unmount
  onCleanup(() => {
    if (es) {
      es.close()
    }
  })

  // Start WebSocket connection when hook is used
  createEffect(() => {
    connectWebSocket()
  })

  return {
    stats,
    connection,
    formatBytes,
    formatUptime,
    getMemoryUsagePercent,
    getCpuUsagePercent,
    getDiskUsagePercent,
    connectWebSocket
  }
}