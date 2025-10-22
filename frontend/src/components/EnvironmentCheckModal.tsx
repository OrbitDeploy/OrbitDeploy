import { createSignal, Component, createEffect } from 'solid-js'
import { apiMutate } from '../lib/apiClient'
import { getContainersApiUrl, getEnvironmentsApiUrl } from '../api/config'

interface EnvironmentCheckData {
  podman: {
    installed: boolean;
    version: string;
    version_valid: boolean;
  };
  caddy: {
    installed: boolean;
    version: string;
  };
  systemd_files?: string[];
  database_containers?: string[];
  missing_in_db?: string[];
  synced_containers?: number;
  errors?: string[];
}

interface EnvironmentCheckResult {
  success: boolean
  data: EnvironmentCheckData
  message: string
}

interface EnvironmentCheckModalProps {
  isOpen: boolean
  onClose: () => void
  onEnvironmentChange?: () => void
}

const EnvironmentCheckModal: Component<EnvironmentCheckModalProps> = (props) => {
  const [loading, setLoading] = createSignal(false)
  const [installing, setInstalling] = createSignal<'podman' | 'caddy' | null>(null)
  const [result, setResult] = createSignal<EnvironmentCheckResult | null>(null)
  const [error, setError] = createSignal('')

  // 根据plan_cn.md，使用统一API配置进行环境检查
  async function checkEnvironment() {
    setLoading(true)
    setError('')
    
    try {
      const data = await apiMutate<EnvironmentCheckData>(getContainersApiUrl('checkEnv'), { method: 'POST' })
      setResult({ success: true, data, message: 'OK' })
      
      // 如果有容器被同步，通知父组件刷新
      if (data.synced_containers && data.synced_containers > 0 && props.onEnvironmentChange) {
        props.onEnvironmentChange()
      }
      
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '环境检查失败'
      setError(errorMessage)
      setResult({ success: false, data: {} as EnvironmentCheckData, message: errorMessage })
    } finally {
      setLoading(false)
    }
  }

  async function installPodman() {
    setInstalling('podman')
    try {
      await apiMutate(getEnvironmentsApiUrl('installPodman'), { method: 'POST' })
      // 安装后重新检查环境
      setTimeout(() => {
        void checkEnvironment()
        setInstalling(null)
      }, 10000) // 10秒后重新检查
    } catch (err) {
      setError('Podman 安装失败: ' + (err instanceof Error ? err.message : '未知错误'))
      setInstalling(null)
    }
  }

  async function installCaddy() {
    setInstalling('caddy')
    try {
      await apiMutate(getEnvironmentsApiUrl('installCaddy'), { method: 'POST' })
      // 安装后重新检查环境
      setTimeout(() => {
        void checkEnvironment()
        setInstalling(null)
      }, 10000) // 10秒后重新检查
    } catch (err) {
      setError('Caddy 安装失败: ' + (err instanceof Error ? err.message : '未知错误'))
      setInstalling(null)
    }
  }

  // 当modal打开时自动检查环境
  createEffect(() => {
    if (props.isOpen && !result()) {
      void checkEnvironment()
    }
  })

  const handleClose = () => {
    setResult(null)
    setError('')
    setLoading(false)
    setInstalling(null)
    props.onClose()
  }

  return (
    <div class={`modal ${props.isOpen ? 'modal-open' : ''}`}>
      <div class="modal-box max-w-4xl">
        <div class="flex items-center justify-between mb-4">
          <h3 class="font-bold text-lg">环境检查结果</h3>
          <button 
            class="btn btn-sm btn-circle btn-ghost"
            onClick={handleClose}
          >
            ✕
          </button>
        </div>

        {error() && (
          <div class="alert alert-error mb-4">
            <span>{error()}</span>
          </div>
        )}

        {loading() && (
          <div class="flex items-center justify-center py-8">
            <span class="loading loading-spinner loading-lg"></span>
          </div>
        )}

        {result() && (
          <div class="space-y-4">
            {/* Podman Status */}
            <div class="card bg-base-200">
              <div class="card-body p-4">
                <h4 class="font-semibold mb-2">Podman 状态</h4>
                <div class="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span class="font-medium">已安装:</span>
                    <span class={`ml-2 badge ${result()!.data.podman.installed ? 'badge-success' : 'badge-error'}`}>
                      {result()!.data.podman.installed ? '已安装' : '未安装'}
                    </span>
                  </div>
                  <div>
                    <span class="font-medium">版本:</span>
                    <span class="ml-2">{result()!.data.podman.version || 'N/A'}</span>
                  </div>
                  <div>
                    <span class="font-medium">版本有效:</span>
                    <span class={`ml-2 badge ${result()!.data.podman.version_valid ? 'badge-success' : 'badge-error'}`}>
                      {result()!.data.podman.version_valid ? '有效' : '无效'}
                    </span>
                  </div>
                </div>
                
                {/* Podman Install Button */}
                {!result()!.data.podman.installed && (
                  <div class="mt-4">
                    <button
                      onClick={() => void installPodman()}
                      disabled={installing() === 'podman'}
                      class="btn btn-primary btn-sm"
                    >
                      {installing() === 'podman' ? (
                        <>
                          <span class="loading loading-spinner loading-sm mr-2"></span>
                          安装中...
                        </>
                      ) : (
                        '安装 Podman'
                      )}
                    </button>
                  </div>
                )}
              </div>
            </div>

            {/* Caddy Status */}
            <div class="card bg-base-200">
              <div class="card-body p-4">
                <h4 class="font-semibold mb-2">Caddy 状态</h4>
                <div class="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span class="font-medium">已安装:</span>
                    <span class={`ml-2 badge ${result()!.data.caddy.installed ? 'badge-success' : 'badge-error'}`}>
                      {result()!.data.caddy.installed ? '已安装' : '未安装'}
                    </span>
                  </div>
                  <div>
                    <span class="font-medium">版本:</span>
                    <span class="ml-2">{result()!.data.caddy.version || 'N/A'}</span>
                  </div>
                </div>
                
                {/* Caddy Install Button */}
                {!result()!.data.caddy.installed && (
                  <div class="mt-4">
                    <button
                      onClick={() => void installCaddy()}
                      disabled={installing() === 'caddy'}
                      class="btn btn-primary btn-sm"
                    >
                      {installing() === 'caddy' ? (
                        <>
                          <span class="loading loading-spinner loading-sm mr-2"></span>
                          安装中...
                        </>
                      ) : (
                        '安装 Caddy'
                      )}
                    </button>
                  </div>
                )}
              </div>
            </div>

            {/* System Files */}
            <div class="card bg-base-200">
              <div class="card-body p-4">
                <h4 class="font-semibold mb-2">系统文件</h4>
                <div class="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span class="font-medium">系统容器:</span>
                    <span class="ml-2">{result()!.data.systemd_files?.length || 0}</span>
                  </div>
                  <div>
                    <span class="font-medium">数据库容器:</span>
                    <span class="ml-2">{result()!.data.database_containers?.length || 0}</span>
                  </div>
                  <div>
                    <span class="font-medium">需要同步:</span>
                    <span class="ml-2">{result()!.data.missing_in_db?.length || 0}</span>
                  </div>
                  <div>
                    <span class="font-medium">已同步:</span>
                    <span class="ml-2 badge badge-info">{result()!.data.synced_containers || 0}</span>
                  </div>
                </div>
                
                {result()!.data.systemd_files && result()!.data.systemd_files.length > 0 && (
                  <div class="mt-3">
                    <span class="font-medium">系统文件:</span>
                    <div class="mt-1 max-h-20 overflow-y-auto">
                      <div class="text-xs text-base-content/70">
                        {result()!.data.systemd_files.join(', ')}
                      </div>
                    </div>
                  </div>
                )}
                
                {result()!.data.missing_in_db && result()!.data.missing_in_db.length > 0 && (
                  <div class="mt-3">
                    <span class="font-medium">已同步到数据库:</span>
                    <div class="mt-1 max-h-20 overflow-y-auto">
                      <div class="text-xs text-base-content/70">
                        {result()!.data.missing_in_db.join(', ')}
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* Errors */}
            {result()!.data.errors && result()!.data.errors.length > 0 && (
              <div class="card bg-error/10 border border-error/20">
                <div class="card-body p-4">
                  <h4 class="font-semibold mb-2 text-error">错误信息</h4>
                  <div class="space-y-1 text-sm">
                    {result()!.data.errors.map((error: string, index: number) => (
                      <div key={index} class="text-error/80">• {error}</div>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {/* Summary */}
            <div class={`alert ${result()!.success ? 'alert-success' : 'alert-warning'}`}>
              <span>{result()!.message}</span>
            </div>
          </div>
        )}

        <div class="modal-action">
          {!loading() && (
            <button
              onClick={() => void checkEnvironment()}
              class="btn btn-primary"
            >
              重新检查
            </button>
          )}
          <button
            onClick={handleClose}
            class="btn btn-ghost"
          >
            关闭
          </button>
        </div>
      </div>
    </div>
  )
}

export default EnvironmentCheckModal