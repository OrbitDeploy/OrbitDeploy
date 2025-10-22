import { getDeploymentsApiUrl, isDev } from '../api/config'

// 部署日志消息结构体 - 与后端保持一致
export interface DeploymentLogMessage {
  deployment_id: number
  message?: string
  timestamp: string
  status?: string
  completed?: boolean
  final?: boolean
}

/**
 * 连接到部署日志的SSE流
 * @param deploymentId 部署ID
 * @param onMessage 收到日志消息时的回调
 * @param onStatusChange 收到状态更新时的回调
 * @param onError 发生错误时的回调
 * @returns EventSource实例，用于关闭连接
 */
export function connectToDeploymentLogsSSE(
  uid: string,
  onMessage: (message: string) => void,
  onStatusChange?: (status: string) => void,
  onError?: (error: string) => void
): EventSource {
  const baseUrl = isDev() ? 'http://localhost:8285' : ''
  
  // 构建SSE URL，包含访问令牌
  const params = new URLSearchParams()
  try {
    const token = (window as any).getAccessToken?.()
    if (token) params.set('access_token', token)
  } catch {}
  
  const sseUrl = `${baseUrl}${getDeploymentsApiUrl({ type: 'logs',  uid })}${params.toString() ? '?' + params.toString() : ''}`

  console.log('Connecting to deployment logs SSE:', sseUrl)
  
  const eventSource = new EventSource(sseUrl)
  
  eventSource.onmessage = (event) => {
    try {
      const data: DeploymentLogMessage = JSON.parse(event.data)
      
      // 处理日志消息
      if (data.message) {
        const timestamp = data.timestamp || new Date().toLocaleTimeString()
        onMessage(`[${timestamp}] ${data.message}`)
      }
      
      // 处理状态更新
      if (data.status && onStatusChange) {
        onStatusChange(data.status)
      }
      
      // 如果是最终状态，关闭连接
      if (data.final) {
        console.log('Deployment finished, closing SSE connection')
        eventSource.close()
      }
    } catch (e) {
      // 如果解析失败，直接显示原始数据
      console.warn('Failed to parse SSE message:', e)
      onMessage(String(event.data))
    }
  }
  
  eventSource.onopen = () => {
    console.log('SSE connection opened for deployment', deploymentId)
  }
  
  eventSource.onerror = (error) => {
    console.error('SSE connection error:', error)
    const errorMessage = `连接错误: ${eventSource.readyState === EventSource.CLOSED ? '连接已关闭' : '连接失败'}`
    if (onError) {
      onError(errorMessage)
    } else {
      onMessage(errorMessage)
    }
  }

  return eventSource
}

/**
 * 格式化部署状态的中文显示
 */
export function formatDeploymentStatus(status: string): string {
  const statusMap: Record<string, string> = {
    'pending': '等待中',
    'running': '运行中',
    'success': '成功',
    'failed': '失败',
    'canceled': '已取消'
  }
  
  return statusMap[status] || status
}

/**
 * 获取部署状态对应的颜色类
 */
export function getDeploymentStatusColor(status: string): string {
  const colorMap: Record<string, string> = {
    'pending': 'text-yellow-600',
    'running': 'text-blue-600',
    'success': 'text-green-600',
    'failed': 'text-red-600',
    'canceled': 'text-gray-600'
  }
  
  return colorMap[status] || 'text-gray-600'
}