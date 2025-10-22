import { Component, createSignal, onCleanup, createEffect, Show, For, onMount } from 'solid-js'
import { createVirtualizer } from '@tanstack/solid-virtual'
import type { Deployment, DeploymentLog } from '../../types/deployment'
import { connectToDeploymentLogsSSE, formatDeploymentStatus, getDeploymentStatusColor } from '../../services/deploymentLogService'
import { getDeploymentLogsDataEndpoint } from '../../api/endpoints/deployments'
import { apiGet } from '../../lib/apiClient'
import { useI18n } from '../../i18n'

interface DeploymentLogsModalProps {
  isOpen: boolean
  onClose: () => void
  deployment: Deployment | null
}

const DeploymentLogsModal: Component<DeploymentLogsModalProps> = (props) => {
  const { t } = useI18n()
  const [logs, setLogs] = createSignal<DeploymentLog[]>([])
  const [currentStatus, setCurrentStatus] = createSignal<string>('')
  const [isConnected, setIsConnected] = createSignal(false)
  const [connectionError, setConnectionError] = createSignal<string>('')
  const [isLoadingOlder, setIsLoadingOlder] = createSignal(false)
  const [hasMoreLogs, setHasMoreLogs] = createSignal(true)
  
  let eventSource: EventSource | null = null
  let logsContainer: HTMLDivElement | undefined
  let scrollParent: HTMLDivElement | undefined

  // 加载初始日志
  const loadInitialLogs = async () => {
    if (!props.deployment) return

    try {
      const endpoint = getDeploymentLogsDataEndpoint(props.deployment.uid, { limit: 200 })
      const data = await apiGet<DeploymentLog[]>(endpoint.url)
      
      if (data) {
        setLogs(data)
        setHasMoreLogs(data.length === 200)
        // 滚动到底部
        setTimeout(scrollToBottom, 100)
      }
    } catch (error) {
      console.error(t('deployment_logs_modal.load_initial_error'), error)
      setConnectionError(`${t('deployment_logs_modal.load_error_prefix')} ${error}`)
    }
  }

  // 加载更早的日志
  const loadOlderLogs = async () => {
    if (!props.deployment || isLoadingOlder() || !hasMoreLogs()) return

    const currentLogs = logs()
    if (currentLogs.length === 0) return

    const oldestLog = currentLogs[0]
    setIsLoadingOlder(true)

    try {
      const endpoint = getDeploymentLogsDataEndpoint(props.deployment.uid, {
        limit: 100,
        before_timestamp: oldestLog.timestamp
      })
      const data = await apiGet<DeploymentLog[]>(endpoint.url)

      if (data && data.length > 0) {
        // 将新加载的日志前插到现有日志数组
        setLogs([...data, ...currentLogs])
        setHasMoreLogs(data.length === 100)
      } else {
        setHasMoreLogs(false)
      }
    } catch (error) {
      console.error(t('deployment_logs_modal.load_older_error'), error)
    } finally {
      setIsLoadingOlder(false)
    }
  }

  // 创建虚拟化器
  const virtualizer = createVirtualizer({
    get count() {
      return logs().length
    },
    getScrollElement: () => scrollParent,
    estimateSize: () => 24, // 每行大约24px高度
    overscan: 10,
  })

  // 滚动到底部
  const scrollToBottom = () => {
    if (scrollParent) {
      scrollParent.scrollTop = scrollParent.scrollHeight
    }
  }

  // 监听滚动事件，检测是否滚动到顶部
  const handleScroll = () => {
    if (!scrollParent) return
    
    // 如果滚动到顶部附近，加载更早的日志
    if (scrollParent.scrollTop < 100 && hasMoreLogs() && !isLoadingOlder()) {
      loadOlderLogs()
    }
  }

  // 连接到SSE
  const connectToLogs = () => {
    if (!props.deployment) return

    setConnectionError('')
    setIsConnected(false)
    setCurrentStatus(props.deployment.status)

    try {
      eventSource = connectToDeploymentLogsSSE(
        props.deployment.uid,
        (message) => {
          // 收到新的日志消息，追加到日志数组
          const newLog: DeploymentLog = {
            id: Date.now(), // 使用时间戳作为临时ID
            timestamp: new Date().toISOString(),
            level: 'INFO',
            source: 'SYSTEM',
            message: message
          }
          setLogs(prev => [...prev, newLog])
          
          // 如果用户在底部，自动滚动
          setTimeout(() => {
            if (scrollParent && scrollParent.scrollHeight - scrollParent.scrollTop - scrollParent.clientHeight < 100) {
              scrollToBottom()
            }
          }, 10)
        },
        (status) => {
          setCurrentStatus(status)
        },
        (error) => {
          setConnectionError(error)
          setIsConnected(false)
        }
      )

      // 监听连接状态
      eventSource.onopen = () => {
        setIsConnected(true)
        setConnectionError('')
      }

    } catch (error) {
      setConnectionError(`${t('deployment_logs_modal.connect_error_prefix')} ${error}`)
    }
  }

  // 断开连接
  const disconnect = () => {
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
    setIsConnected(false)
  }

  // 清除日志
  const clearLogs = () => {
    setLogs([])
  }

  // 复制日志到剪贴板
  const copyLogs = async () => {
    try {
      const logText = logs().map(log => 
        `[${new Date(log.timestamp).toLocaleString()}] [${log.level}] ${log.message}`
      ).join('\n')
      await navigator.clipboard.writeText(logText)
    } catch (error) {
      console.error(t('deployment_logs_modal.copy_error'), error)
    }
  }

  // 下载日志文件
  const downloadLogs = () => {
    if (!props.deployment) return
    
    const logText = logs().map(log => 
      `[${new Date(log.timestamp).toLocaleString()}] [${log.level}] ${log.message}`
    ).join('\n')
    const blob = new Blob([logText], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `deployment-${props.deployment.uid}-logs.txt`
    a.click()
    URL.revokeObjectURL(url)
  }

  // 清理资源
  onCleanup(() => {
    disconnect()
  })

  // 监听模态框打开状态变化
  createEffect(() => {
    if (props.isOpen && props.deployment) {
      loadInitialLogs()
      connectToLogs()
    } else {
      disconnect()
      setLogs([])
    }
  })

  // 获取日志级别颜色
  const getLevelColor = (level: string) => {
    switch (level.toUpperCase()) {
      case 'ERROR': return 'text-error'
      case 'WARN': return 'text-warning'
      case 'INFO': return 'text-info'
      case 'DEBUG': return 'text-base-content/70'
      default: return 'text-base-content'
    }
  }

  return (
    <Show when={props.isOpen}>
      <div class="modal modal-open">
        <div class="modal-box w-11/12 max-w-5xl h-5/6 flex flex-col">
          {/* 模态框头部 */}
          <div class="flex justify-between items-center mb-4">
            <div>
              <h3 class="font-bold text-lg">
                {t('deployment_logs_modal.title')} #{props.deployment?.uid}
              </h3>
              <div class="flex items-center gap-2 mt-1">
                <span class={`badge badge-sm ${getDeploymentStatusColor(currentStatus())}`}>
                  {formatDeploymentStatus(currentStatus())}
                </span>
                <Show when={isConnected()}>
                  <span class="badge badge-success badge-sm">{t('deployment_logs_modal.status_connected')}</span>
                </Show>
                <Show when={connectionError()}>
                  <span class="badge badge-error badge-sm" title={connectionError()}>
                    {t('deployment_logs_modal.status_error')}
                  </span>
                </Show>
              </div>
            </div>
            
            {/* 操作按钮 */}
            <div class="flex gap-2">
              <div class="tooltip" data-tip={t('deployment_logs_modal.tooltip_reconnect')}>
                <button
                  class="btn btn-sm btn-outline"
                  onClick={connectToLogs}
                  disabled={isConnected()}
                >
                  🔄
                </button>
              </div>
              <div class="tooltip" data-tip={t('deployment_logs_modal.tooltip_clear')}>
                <button
                  class="btn btn-sm btn-outline"
                  onClick={clearLogs}
                >
                  🗑️
                </button>
              </div>
              <div class="tooltip" data-tip={t('deployment_logs_modal.tooltip_copy')}>
                <button
                  class="btn btn-sm btn-outline"
                  onClick={copyLogs}
                >
                  📋
                </button>
              </div>
              <div class="tooltip" data-tip={t('deployment_logs_modal.tooltip_download')}>
                <button
                  class="btn btn-sm btn-outline"
                  onClick={downloadLogs}
                >
                  📥
                </button>
              </div>
              <button class="btn btn-sm" onClick={props.onClose}>
                {t('common.close')}
              </button>
            </div>
          </div>

          {/* 日志内容区域 - 虚拟化列表 */}
          <div class="flex-1 overflow-hidden">
            <div
              ref={scrollParent}
              class="h-full w-full bg-base-300 rounded-lg p-4 overflow-auto"
              onScroll={handleScroll}
            >
              <Show when={isLoadingOlder()}>
                <div class="text-center py-2">
                  <span class="loading loading-spinner loading-sm"></span>
                  <span class="ml-2 text-sm">{t('deployment_logs_modal.loading_older')}</span>
                </div>
              </Show>
              
              <div
                style={{
                  height: `${virtualizer.getTotalSize()}px`,
                  width: '100%',
                  position: 'relative',
                }}
              >
                <For each={virtualizer.getVirtualItems()}>
                  {(virtualRow) => {
                    const log = logs()[virtualRow.index]
                    return (
                      <div
                        style={{
                          position: 'absolute',
                          top: 0,
                          left: 0,
                          width: '100%',
                          transform: `translateY(${virtualRow.start}px)`,
                        }}
                        class="font-mono text-sm"
                      >
                        <span class="text-base-content/50">
                          [{new Date(log.timestamp).toLocaleTimeString()}]
                        </span>
                        <span class={`ml-2 ${getLevelColor(log.level)}`}>
                          [{log.level}]
                        </span>
                        <span class="ml-2 text-base-content">
                          {log.message}
                        </span>
                      </div>
                    )
                  }}
                </For>
              </div>

              <Show when={logs().length === 0}>
                <div class="text-center py-8 text-base-content/70">
                  {t('deployment_logs_modal.no_logs')}
                </div>
              </Show>
            </div>
          </div>

          {/* 底部状态栏 */}
          <div class="mt-4 text-sm text-base-content/70">
            <Show when={props.deployment}>
              <div class="flex justify-between items-center">
                <span>
                  {t('deployment_logs_modal.footer_image')} {props.deployment?.imageName || t('deployment_logs_modal.footer_unknown')}
                </span>
                <span>
                  {t('deployment_logs_modal.footer_log_count', { count: logs().length })}
                </span>
              </div>
            </Show>
          </div>
        </div>
        
        {/* 模态框背景 */}
        <div class="modal-backdrop" onClick={props.onClose}></div>
      </div>
    </Show>
  )
}

export default DeploymentLogsModal