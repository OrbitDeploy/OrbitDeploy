import { createSignal, createEffect, onCleanup } from 'solid-js'
import { buildApiUrl, API_ENDPOINTS } from '../api/config'

interface RunningDeploymentSummary {
  total_running: number
  last_updated: string
  deployments: RunningDeploymentDetail[]
}

interface RunningDeploymentDetail {
  app_name: string
  version: string
}

interface ConnectionStatus {
  status: 'connecting' | 'connected' | 'reconnecting' | 'disconnected'
  lastUpdate?: number
}

export function useRunningDeployments() {
  const [summary, setSummary] = createSignal<RunningDeploymentSummary | null>(null)
  const [connection, setConnection] = createSignal<ConnectionStatus>({ status: 'connecting' })
  
  let es: EventSource | null = null

  // Format last updated time
  const formatLastUpdated = (timestamp: string): string => {
    const date = new Date(timestamp)
    return date.toLocaleString('zh-CN')
  }

  // Connect to SSE
  const connectSSE = () => {
    try {
      setConnection({ status: 'connecting' })

      const params = new URLSearchParams()
      try {
        const token = (window as any).getAccessToken?.()
        if (token) params.set('access_token', token)
      } catch {}
      // 修复：去掉/api前缀，避免重复
      const sseUrl = `${buildApiUrl('/system/running-deployments')}${params.toString() ? '?' + params.toString() : ''}`

      es = new EventSource(sseUrl)

      es.onopen = () => {
        console.log('Running deployments SSE connected')
        setConnection({ status: 'connected', lastUpdate: Date.now() })
      }

      es.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as RunningDeploymentSummary
          setSummary(data)
          setConnection({ status: 'connected', lastUpdate: Date.now() })
        } catch (err) {
          console.error('Failed to parse running deployments SSE message:', err)
        }
      }

      es.addEventListener('running-deployments', (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data) as RunningDeploymentSummary
          setSummary(data)
          setConnection({ status: 'connected', lastUpdate: Date.now() })
        } catch (err) {
          console.error('Failed to parse running-deployments SSE message:', err)
        }
      })

      es.onerror = (error) => {
        console.error('Running deployments SSE error:', error)
        setConnection({ status: 'reconnecting' })
      }
    } catch (err) {
      console.error('Failed to connect running deployments SSE:', err)
      setConnection({ status: 'disconnected' })
    }
  }

  onCleanup(() => {
    if (es) {
      es.close()
    }
  })

  createEffect(() => {
    connectSSE()
  })

  return {
    summary,
    connection,
    formatLastUpdated
  }
}
