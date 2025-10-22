import { createSignal, Component, onMount } from 'solid-js'
import { apiGet, apiMutate } from '../lib/apiClient'
import { getEnvironmentsApiUrl } from '../api/config'

interface EnvironmentStatus {
  podman: {
    installed: boolean
    version: string
    version_valid: boolean
    message: string
  }
  caddy: {
    installed: boolean
    version: string
    message: string
  }
  overall_status: 'ready' | 'partial' | 'missing'
}

interface ApiResponse<T> {
  success: boolean
  data?: T
  message?: string
}

const EnvironmentCheckPanel: Component = () => {
  const [status, setStatus] = createSignal<EnvironmentStatus | null>(null)
  const [loading, setLoading] = createSignal(false)
  const [installing, setInstalling] = createSignal<'podman' | 'caddy' | null>(null)

  async function loadEnvironmentStatus() {
    try {
      const data = await apiGet<EnvironmentStatus>(getEnvironmentsApiUrl('check'))
      setStatus(data)
    } catch (e) {
      console.error('Failed to load environment status', e)
    }
  }
  onMount(loadEnvironmentStatus)


  const installPodman = async () => {
    setInstalling('podman')
    try {
      await apiMutate(getEnvironmentsApiUrl('installPodman'), { method: 'POST' })
      // Poll for status updates
      setTimeout(() => {
        void loadEnvironmentStatus()
        setInstalling(null)
      }, 10000) // Check again after 10 seconds
    } catch (err) {
      console.error('Failed to install Podman:', err)
      setInstalling(null)
    }
  }

  const installCaddy = async () => {
    setInstalling('caddy')
    try {
      await apiMutate(getEnvironmentsApiUrl('installCaddy'), { method: 'POST' })
      // Poll for status updates
      setTimeout(() => {
        void loadEnvironmentStatus()
        setInstalling(null)
      }, 10000) // Check again after 10 seconds
    } catch (err) {
      console.error('Failed to install Caddy:', err)
      setInstalling(null)
    }
  }

  const refresh = () => {
    void loadEnvironmentStatus()
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'ready':
        return (
          <svg class="w-5 h-5 text-success" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
          </svg>
        )
      case 'partial':
        return (
          <svg class="w-5 h-5 text-warning" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
          </svg>
        )
      default:
        return (
          <svg class="w-5 h-5 text-error" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
          </svg>
        )
    }
  }

  return (
    <div class="card bg-base-100 shadow-lg mb-6">
      <div class="card-body">
        <div class="flex items-center justify-between mb-4">
          <h2 class="card-title text-primary">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-6 h-6 stroke-current">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            部署环境检查
          </h2>
          <button 
            class="btn btn-sm btn-ghost"
            onClick={refresh}
            disabled={loading()}
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="w-4 h-4 stroke-current">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            刷新
          </button>
        </div>

        {status() ? (
          <div class="space-y-4">
            {/* Overall Status */}
            <div class="alert alert-info">
              <div class="flex items-center gap-3">
                {getStatusIcon(status()!.overall_status)}
                <span class="font-medium">
                  {status()!.overall_status === 'ready' ? '环境准备就绪' :
                   status()!.overall_status === 'partial' ? '环境部分准备完成' : '环境需要配置'}
                </span>
              </div>
            </div>

            {/* Podman Status */}
            <div class="flex items-center justify-between p-4 bg-base-50 rounded-lg">
              <div class="flex items-center gap-3">
                <div class={`w-3 h-3 rounded-full ${
                  status()!.podman.installed && status()!.podman.version_valid ? 'bg-success' : 
                  status()!.podman.installed ? 'bg-warning' : 'bg-error'
                }`}></div>
                <div>
                  <div class="font-medium">Podman</div>
                  <div class="text-sm text-base-content/70">
                    {status()!.podman.message}
                  </div>
                  {status()!.podman.installed && (
                    <div class="text-xs text-base-content/50">
                      版本: {status()!.podman.version}
                    </div>
                  )}
                </div>
              </div>
              <div>
                {!status()!.podman.installed ? (
                  <button 
                    class="btn btn-primary btn-sm"
                    onClick={installPodman}
                    disabled={installing() === 'podman'}
                  >
                    {installing() === 'podman' ? (
                      <>
                        <span class="loading loading-spinner loading-sm"></span>
                        安装中...
                      </>
                    ) : (
                      '安装 Podman'
                    )}
                  </button>
                ) : !status()!.podman.version_valid ? (
                  <div class="text-warning text-sm">
                    请先卸载后安装最新 Podman
                  </div>
                ) : (
                  <div class="badge badge-success">已安装</div>
                )}
              </div>
            </div>

            {/* Caddy Status */}
            <div class="flex items-center justify-between p-4 bg-base-50 rounded-lg">
              <div class="flex items-center gap-3">
                <div class={`w-3 h-3 rounded-full ${
                  status()!.caddy.installed ? 'bg-success' : 'bg-error'
                }`}></div>
                <div>
                  <div class="font-medium">Caddy</div>
                  <div class="text-sm text-base-content/70">
                    {status()!.caddy.message}
                  </div>
                  {status()!.caddy.installed && (
                    <div class="text-xs text-base-content/50">
                      版本: {status()!.caddy.version}
                    </div>
                  )}
                </div>
              </div>
              <div>
                {!status()!.caddy.installed ? (
                  <button 
                    class="btn btn-primary btn-sm"
                    onClick={installCaddy}
                    disabled={installing() === 'caddy'}
                  >
                    {installing() === 'caddy' ? (
                      <>
                        <span class="loading loading-spinner loading-sm"></span>
                        安装中...
                      </>
                    ) : (
                      '安装 Caddy'
                    )}
                  </button>
                ) : (
                  <div class="badge badge-success">已安装</div>
                )}
              </div>
            </div>
          </div>
        ) : (
          <div class="flex items-center justify-center py-8">
            <span class="loading loading-spinner loading-lg"></span>
          </div>
        )}
      </div>
    </div>
  )
}

export default EnvironmentCheckPanel