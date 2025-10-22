import { createSignal, createEffect, Component, For } from 'solid-js'
import { createFetch } from '@solid-primitives/fetch'
import { extractHostPortFromQuadlet, stripProtocolFromDomain, isValidDomain } from '../lib/utils'
import { useI18n } from '../i18n'

interface DomainStatus {
  domain: string
  configured: boolean
  proxy_target: string
  message: string
}

interface ContainerDomainStatus {
  container_id: number
  container_name: string
  quadlet_file: string
  domains: DomainStatus[]
}

interface ApiResponse<T> {
  success: boolean
  data?: T
  message?: string
}

interface DomainManagementModalProps {
  containerId: number
  isOpen: boolean
  onClose: () => void
  onDomainChange?: () => void
}

const DomainManagementModal: Component<DomainManagementModalProps> = (props) => {
  const { t } = useI18n()
  const [domainStatus, setDomainStatus] = createSignal<ContainerDomainStatus | null>(null)
  const [loading, setLoading] = createSignal(false)
  const [newDomain, setNewDomain] = createSignal('')
  const [newPort, setNewPort] = createSignal('')
  const [newHost, setNewHost] = createSignal('localhost')
  const [error, setError] = createSignal('')

  const [domainsResponse, { refetch: refetchDomains }] = createFetch<ApiResponse<ContainerDomainStatus>>(
    () => props.isOpen ? `/api/containers/${props.containerId}/domains` : null
  )

  createEffect(() => {
    if (domainsResponse()) {
      const response = domainsResponse()
      if (response && response.success && response.data) {
        setDomainStatus(response.data)
        
        // Always populate port from Quadlet file for current container
        if (response.data.quadlet_file) {
          const extractedPort = extractHostPortFromQuadlet(response.data.quadlet_file)
          if (extractedPort) {
            setNewPort(extractedPort.toString())
          }
        }
      }
    }
  })

  // Reset form when container changes
  createEffect(() => {
    if (props.isOpen) {
      setNewDomain('')
      setNewHost('localhost')
      setError('')
    }
  })

  const addDomain = async () => {
    if (!newDomain().trim() || !newPort().trim()) {
      setError(t('containers.domain_management.domain_required'))
      return
    }

    // Strip protocol and validate domain
    const cleanDomain = stripProtocolFromDomain(newDomain().trim())
    if (!cleanDomain || !isValidDomain(cleanDomain)) {
      setError(t('containers.domain_management.invalid_domain'))
      return
    }

    setLoading(true)
    setError('')

    try {
      const response = await fetch(`/api/containers/${props.containerId}/domains/manage`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          domain: cleanDomain, // Use cleaned domain
          port: parseInt(newPort()),
          host: newHost().trim() || 'localhost',
          action: 'add'
        })
      })

      const result = await response.json()
      
      if (response.ok && result.success) {
        setNewDomain('')
        setNewPort('')
        setNewHost('localhost')
        refetchDomains()
        // Notify parent component that domains have changed
        props.onDomainChange?.()
      } else {
        setError(result.message || t('containers.domain_management.add_failed'))
      }
    } catch (err) {
      setError(t('containers.domain_management.add_error'))
      console.error('Add domain error:', err)
    } finally {
      setLoading(false)
    }
  }

  const removeDomain = async (domain: string) => {
    if (!confirm(t('containers.domain_management.confirm_remove').replace('{domain}', domain))) {
      return
    }

    setLoading(true)
    setError('')

    try {
      const response = await fetch(`/api/containers/${props.containerId}/domains/manage`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          domain: domain,
          action: 'remove'
        })
      })

      const result = await response.json()
      
      if (response.ok && result.success) {
        refetchDomains()
        // Notify parent component that domains have changed
        props.onDomainChange?.()
      } else {
        setError(result.message || t('containers.domain_management.remove_failed'))
      }
    } catch (err) {
      setError(t('containers.domain_management.remove_error'))
      console.error('Remove domain error:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleClose = () => {
    setError('')
    setNewDomain('')
    setNewPort('')
    setNewHost('localhost')
    props.onClose()
  }

  return (
    <div class={`modal ${props.isOpen ? 'modal-open' : ''}`}>
      <div class="modal-box w-11/12 max-w-2xl">
        <div class="flex items-center justify-between mb-4">
          <h3 class="font-bold text-lg">{t('containers.domain_management.title')}</h3>
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

        {/* Add New Domain */}
        <div class="card bg-base-50 mb-6">
          <div class="card-body">
            <h4 class="font-medium mb-3">{t('containers.domain_management.add_new')}</h4>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div class="form-control">
                <label class="label">
                  <span class="label-text">{t('containers.domain_management.domain_field')}</span>
                </label>
                <input
                  type="text"
                  placeholder={t('containers.domain_management.domain_placeholder')}
                  value={newDomain()}
                  onInput={(e) => setNewDomain(e.target.value)}
                  class="input input-bordered input-sm"
                  disabled={loading()}
                />
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">{t('containers.domain_management.port_field')}</span>
                </label>
                <input
                  type="number"
                  placeholder={t('containers.domain_management.port_placeholder')}
                  value={newPort()}
                  class="input input-bordered input-sm"
                  disabled={true}
                  title="端口号从容器配置自动读取"
                />
                <label class="label">
                  <span class="label-text-alt text-info">端口号从容器信息自动读取，不可修改</span>
                </label>
              </div>
              <div class="form-control">
                <label class="label">
                  <span class="label-text">{t('containers.domain_management.host_field')}</span>
                </label>
                <input
                  type="text"
                  placeholder={t('containers.domain_management.host_placeholder')}
                  value={newHost()}
                  onInput={(e) => setNewHost(e.target.value)}
                  class="input input-bordered input-sm"
                  disabled={loading()}
                />
              </div>
            </div>
            <div class="card-actions justify-end mt-3">
              <button
                class="btn btn-primary btn-sm"
                onClick={addDomain}
                disabled={loading() || !newDomain().trim() || !newPort().trim()}
              >
                {loading() ? (
                  <>
                    <span class="loading loading-spinner loading-sm"></span>
                    {t('containers.domain_management.adding')}
                  </>
                ) : (
                  t('containers.domain_management.add_button')
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Current Domains */}
        <div class="card bg-base-100">
          <div class="card-body">
            <h4 class="font-medium mb-3">{t('containers.domain_management.current_domain_config')}</h4>
            {domainStatus() ? (
              domainStatus()!.domains.length > 0 ? (
                <div class="space-y-3">
                  <For each={domainStatus()!.domains}>
                    {(domain) => (
                      <div class="flex items-center justify-between p-3 bg-base-50 rounded-lg">
                        <div class="flex items-center gap-3">
                          <div class={`w-3 h-3 rounded-full ${
                            domain.configured ? 'bg-success' : 'bg-warning'
                          }`}></div>
                          <div>
                            <div class="font-medium">{domain.domain}</div>
                            <div class="text-sm text-base-content/70">
                              {domain.message}
                            </div>
                            {domain.proxy_target && (
                              <div class="text-xs text-base-content/50">
                                {t('containers.domain_management.proxy_to')}: {domain.proxy_target}
                              </div>
                            )}
                          </div>
                        </div>
                        <div class="flex items-center gap-2">
                          {domain.configured ? (
                            <div class="badge badge-success badge-sm">{t('containers.domain_management.configured')}</div>
                          ) : (
                            <div class="badge badge-warning badge-sm">{t('containers.domain_management.not_configured')}</div>
                          )}
                          <button
                            class="btn btn-error btn-xs"
                            onClick={() => removeDomain(domain.domain)}
                            disabled={loading()}
                          >
                            {t('containers.domain_management.remove_button')}
                          </button>
                        </div>
                      </div>
                    )}
                  </For>
                </div>
              ) : (
                <div class="text-center py-8 text-base-content/60">
                  {t('containers.domain_management.no_domains_configured')}
                </div>
              )
            ) : (
              <div class="flex items-center justify-center py-8">
                <span class="loading loading-spinner loading-lg"></span>
              </div>
            )}
          </div>
        </div>

        <div class="modal-action">
          <button class="btn" onClick={handleClose}>{t('containers.actions.close')}</button>
        </div>
      </div>
    </div>
  )
}

export default DomainManagementModal