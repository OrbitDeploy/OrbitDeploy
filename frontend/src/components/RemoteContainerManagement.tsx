import { createSignal, createEffect, For, Show, onCleanup } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n } from '../i18n'
import type { RemoteContainer, PodmanConnection, SSHHost, RemoteContainerManagementProps } from '../types/remote'

const RemoteContainerManagement: Component<RemoteContainerManagementProps> = (props) => {
  const { t } = useI18n()
  
  // State management
  const [selectedHost, setSelectedHost] = createSignal<SSHHost | null>(null)
  const [containers, setContainers] = createSignal<RemoteContainer[]>([])
  const [connections, setConnections] = createSignal<PodmanConnection[]>([])
  const [loading, setLoading] = createSignal(false)
  const [error, setError] = createSignal('')
  const [success, setSuccess] = createSignal('')
  
  // Modal states
  const [showContainerLogs, setShowContainerLogs] = createSignal(false)
  const [selectedContainer, setSelectedContainer] = createSignal<RemoteContainer | null>(null)
  const [containerLogs, setContainerLogs] = createSignal('')
  
  // Auto refresh
  const [autoRefresh, setAutoRefresh] = createSignal(false)
  let refreshInterval: number | null = null

  // Load remote containers for selected host
  const loadRemoteContainers = async (host: SSHHost) => {
    if (!host) return
    
    setLoading(true)
    setError('')
    
    try {
      const response = await fetch(`/api/ssh-hosts/${host.uid}/containers`)
      const data = await response.json()
      
      if (data.success) {
        setContainers(data.data || [])
      } else {
        setError(data.message || 'Failed to load remote containers')
        setContainers([])
      }
    } catch (err) {
      setError('Failed to connect to remote host')
      setContainers([])
    } finally {
      setLoading(false)
    }
  }

  // Load Podman connections for selected host
  const loadPodmanConnections = async (host: SSHHost) => {
    if (!host) return
    
    try {
      const response = await fetch(`/api/ssh-hosts/${host.uid}/podman/connections`)
      const data = await response.json()
      
      if (data.success) {
        setConnections(data.data || [])
      } else {
        setConnections([])
      }
    } catch (err) {
      setConnections([])
    }
  }

  // Control remote container (start/stop/restart)
  const controlContainer = async (container: RemoteContainer, action: string) => {
    setLoading(true)
    setError('')
    
    try {
      const response = await fetch(`/api/ssh-hosts/${container.host_id}/containers/${container.names[0]}/${action}`, {
        method: 'POST'
      })
      const data = await response.json()
      
      if (data.success) {
        setSuccess(`Container ${action}ed successfully`)
        // Reload containers to update status
        const host = selectedHost()
        if (host) {
          await loadRemoteContainers(host)
        }
      } else {
        setError(data.message || `Failed to ${action} container`)
      }
    } catch (err) {
      setError(`Failed to ${action} container`)
    } finally {
      setLoading(false)
    }
  }

  // View container logs
  const viewContainerLogs = async (container: RemoteContainer) => {
    setSelectedContainer(container)
    setContainerLogs('')
    setShowContainerLogs(true)
    
    try {
      const response = await fetch(`/api/ssh-hosts/${container.host_id}/containers/${container.names[0]}/logs`)
      const data = await response.json()
      
      if (data.success && data.data?.logs) {
        setContainerLogs(data.data.logs)
      } else {
        setContainerLogs('No logs available or failed to fetch logs')
      }
    } catch (err) {
      setContainerLogs('Failed to fetch container logs')
    }
  }

  // Handle host selection
  const selectHost = (host: SSHHost) => {
    setSelectedHost(host)
    setContainers([])
    setConnections([])
    loadRemoteContainers(host)
    loadPodmanConnections(host)
  }

  // Auto refresh effect
  createEffect(() => {
    if (autoRefresh()) {
      refreshInterval = window.setInterval(() => {
        const host = selectedHost()
        if (host) {
          loadRemoteContainers(host)
        }
      }, 5000) // Refresh every 5 seconds
    } else {
      if (refreshInterval) {
        clearInterval(refreshInterval)
        refreshInterval = null
      }
    }
  })

  // Cleanup on unmount
  onCleanup(() => {
    if (refreshInterval) {
      clearInterval(refreshInterval)
    }
  })

  // Get container status badge class
  const getStatusBadgeClass = (status: string) => {
    if (status.toLowerCase().includes('running')) return 'badge-success'
    if (status.toLowerCase().includes('stopped') || status.toLowerCase().includes('exited')) return 'badge-error'
    if (status.toLowerCase().includes('paused')) return 'badge-warning'
    return 'badge-neutral'
  }

  // Format date
  const formatDate = (dateStr: string) => {
    try {
      return new Date(dateStr).toLocaleString()
    } catch {
      return dateStr
    }
  }

  return (
    <div class="space-y-6">
      {/* Header */}
      <div class="flex justify-between items-center">
        <h3 class="text-xl font-semibold">Remote Container Management</h3>
        <div class="flex gap-2 items-center">
          <label class="label cursor-pointer">
            <span class="label-text mr-2">Auto Refresh</span>
            <input 
              type="checkbox" 
              class="toggle toggle-primary" 
              checked={autoRefresh()}
              onChange={(e) => setAutoRefresh(e.target.checked)}
            />
          </label>
          <button 
            class="btn btn-sm btn-primary"
            onClick={() => {
              const host = selectedHost()
              if (host) {
                loadRemoteContainers(host)
                loadPodmanConnections(host)
              }
            }}
            disabled={loading() || !selectedHost()}
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-4 h-4">
              <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
            </svg>
            Refresh
          </button>
        </div>
      </div>

      {/* Error/Success Messages */}
      <Show when={error()}>
        <div class="alert alert-error">
          <span>{error()}</span>
          <button class="btn btn-sm btn-ghost" onClick={() => setError('')}>×</button>
        </div>
      </Show>

      <Show when={success()}>
        <div class="alert alert-success">
          <span>{success()}</span>
          <button class="btn btn-sm btn-ghost" onClick={() => setSuccess('')}>×</button>
        </div>
      </Show>

      {/* Host Selection */}
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h4 class="card-title">Select SSH Host</h4>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <For each={props.hosts}>
              {(host) => (
                <div 
                  class={`card bg-base-200 cursor-pointer transition-all hover:bg-base-300 ${
                    selectedHost()?.uid === host.uid ? 'ring-2 ring-primary' : ''
                  }`}
                  onClick={() => selectHost(host)}
                >
                  <div class="card-body p-4">
                    <h5 class="font-medium">{host.name}</h5>
                    <p class="text-sm text-base-content/70">{host.addr}:{host.port}</p>
                    <p class="text-sm text-base-content/70">User: {host.user}</p>
                    <Show when={host.description}>
                      <p class="text-xs text-base-content/60">{host.description}</p>
                    </Show>
                  </div>
                </div>
              )}
            </For>
          </div>
        </div>
      </div>

      {/* Podman Connections */}
      <Show when={selectedHost() && connections().length > 0}>
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <h4 class="card-title">Podman System Connections</h4>
            <div class="overflow-x-auto">
              <table class="table table-sm">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>URI</th>
                    <th>Identity</th>
                    <th>Default</th>
                  </tr>
                </thead>
                <tbody>
                  <For each={connections()}>
                    {(conn) => (
                      <tr>
                        <td class="font-medium">{conn.name}</td>
                        <td class="font-mono text-sm">{conn.uri}</td>
                        <td class="text-sm">{conn.identity}</td>
                        <td>
                          <Show when={conn.default}>
                            <span class="badge badge-primary badge-sm">Default</span>
                          </Show>
                        </td>
                      </tr>
                    )}
                  </For>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </Show>

      {/* Remote Containers */}
      <Show when={selectedHost()}>
        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <div class="flex justify-between items-center mb-4">
              <h4 class="card-title">
                Remote Containers
                <Show when={selectedHost()}>
                  <span class="badge badge-neutral">{selectedHost()?.name}</span>
                </Show>
              </h4>
              <div class="text-sm text-base-content/70">
                {containers().length} containers found
              </div>
            </div>
            
            <Show when={loading()}>
              <div class="flex justify-center p-6">
                <span class="loading loading-spinner loading-lg"></span>
              </div>
            </Show>

            <Show when={!loading() && containers().length === 0}>
              <div class="text-center p-6 text-base-content/70">
                No containers found on this host
              </div>
            </Show>

            <Show when={!loading() && containers().length > 0}>
              <div class="overflow-x-auto">
                <table class="table">
                  <thead>
                    <tr>
                      <th>Name</th>
                      <th>Image</th>
                      <th>Status</th>
                      <th>Ports</th>
                      <th>Created</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    <For each={containers()}>
                      {(container) => (
                        <tr>
                          <td>
                            <div class="font-medium">{container.names[0] || container.uid.substring(0, 12)}</div>
                            <div class="text-xs text-base-content/60">ID: {container.uid.substring(0, 12)}</div>
                          </td>
                          <td>
                            <div class="font-mono text-sm">{container.image}</div>
                          </td>
                          <td>
                            <span class={`badge ${getStatusBadgeClass(container.status)}`}>
                              {container.status}
                            </span>
                          </td>
                          <td class="font-mono text-sm">{container.ports}</td>
                          <td class="text-sm">{formatDate(container.created_at)}</td>
                          <td>
                            <div class="flex gap-1">
                              <button 
                                class="btn btn-xs btn-success"
                                onClick={() => controlContainer(container, 'start')}
                                disabled={loading() || container.status.toLowerCase().includes('running')}
                                title="Start Container"
                              >
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-3 h-3">
                                  <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.348a1.125 1.125 0 010 1.971l-11.54 6.347a1.125 1.125 0 01-1.667-.985V5.653z" />
                                </svg>
                              </button>
                              
                              <button 
                                class="btn btn-xs btn-error"
                                onClick={() => controlContainer(container, 'stop')}
                                disabled={loading() || !container.status.toLowerCase().includes('running')}
                                title="Stop Container"
                              >
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-3 h-3">
                                  <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 7.5A2.25 2.25 0 017.5 5.25h9a2.25 2.25 0 012.25 2.25v9a2.25 2.25 0 01-2.25 2.25h-9a2.25 2.25 0 01-2.25-2.25v-9z" />
                                </svg>
                              </button>
                              
                              <button 
                                class="btn btn-xs btn-warning"
                                onClick={() => controlContainer(container, 'restart')}
                                disabled={loading()}
                                title="Restart Container"
                              >
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-3 h-3">
                                  <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
                                </svg>
                              </button>
                              
                              <button 
                                class="btn btn-xs btn-info"
                                onClick={() => viewContainerLogs(container)}
                                disabled={loading()}
                                title="View Logs"
                              >
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-3 h-3">
                                  <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0-1.125-.504-1.125-1.125V11.25a9 9 0 00-9-9z" />
                                </svg>
                              </button>
                            </div>
                          </td>
                        </tr>
                      )}
                    </For>
                  </tbody>
                </table>
              </div>
            </Show>
          </div>
        </div>
      </Show>

      {/* Container Logs Modal */}
      <Show when={showContainerLogs()}>
        <div class="modal modal-open">
          <div class="modal-box max-w-4xl h-3/4">
            <h3 class="font-bold text-lg mb-4">
              Container Logs - {selectedContainer()?.names[0]}
            </h3>
            
            <div class="bg-black text-green-400 font-mono text-sm p-4 rounded h-full max-h-96 overflow-y-auto">
              <pre class="whitespace-pre-wrap">{containerLogs()}</pre>
            </div>

            <div class="modal-action">
              <button 
                class="btn"
                onClick={() => {
                  setShowContainerLogs(false)
                  setSelectedContainer(null)
                  setContainerLogs('')
                }}
              >
                Close
              </button>
            </div>
          </div>
        </div>
      </Show>
    </div>
  )
}

export default RemoteContainerManagement