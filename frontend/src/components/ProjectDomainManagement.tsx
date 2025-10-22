import { createSignal, createEffect, Component, For } from 'solid-js'
import { createFetch } from '@solid-primitives/fetch'
import { stripProtocolFromDomain, isValidDomain } from '../lib/utils'
import { useI18n } from '../i18n'
import type { Project } from '../types/project'

interface ProjectDomainStatus {
  domain: string
  port: number
  host: string
  configured: boolean
  proxy_target: string
  message: string
}

interface ProjectDomainManagementStatus {
  project_id: number
  project_name: string
  domains: ProjectDomainStatus[]
}

interface ApiResponse<T> {
  success: boolean
  data?: T
  message?: string
}

interface ProjectDomainManagementProps {
  project: Project
  onDomainChange?: () => void
}

const ProjectDomainManagement: Component<ProjectDomainManagementProps> = (props) => {
  const { t } = useI18n()
  const [domainStatus, setDomainStatus] = createSignal<ProjectDomainManagementStatus | null>(null)
  const [loading, setLoading] = createSignal(false)
  const [newDomain, setNewDomain] = createSignal('')
  const [newPort, setNewPort] = createSignal('8080')
  const [newHost, setNewHost] = createSignal('localhost')
  const [error, setError] = createSignal('')

  const [domainsResponse, { refetch: refetchDomains }] = createFetch<ApiResponse<ProjectDomainManagementStatus>>(
    () => `/api/projects/${props.project.uid}/domains`
  )

  createEffect(() => {
    const response = domainsResponse()
    if (response?.success && response.data) {
      setDomainStatus(response.data)
    }
  })

  const addDomain = async () => {
    if (!newDomain().trim()) {
      setError('Domain is required')
      return
    }

    // Strip protocol and validate domain
    const cleanDomain = stripProtocolFromDomain(newDomain().trim())
    if (!cleanDomain || !isValidDomain(cleanDomain)) {
      setError('Invalid domain format')
      return
    }

    setLoading(true)
    setError('')

    try {
      const response = await fetch(`/api/projects/${props.project.uid}/domains/manage`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          domain: cleanDomain,
          port: parseInt(newPort()),
          host: newHost().trim() || 'localhost',
          action: 'add'
        })
      })

      const result = await response.json()
      
      if (response.ok && result.success) {
        setNewDomain('')
        setNewPort('8080')
        setNewHost('localhost')
        refetchDomains()
        props.onDomainChange?.()
      } else {
        setError(result.message || 'Failed to add domain')
      }
    } catch (err) {
      setError('Failed to add domain')
    } finally {
      setLoading(false)
    }
  }

  const removeDomain = async (domain: string) => {
    setLoading(true)
    setError('')

    try {
      const response = await fetch(`/api/projects/${props.project.uid}/domains/manage`, {
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
        props.onDomainChange?.()
      } else {
        setError(result.message || 'Failed to remove domain')
      }
    } catch (err) {
      setError('Failed to remove domain')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div class="space-y-4">
      <h3 class="text-lg font-semibold">Domain Management</h3>
      
      <div class="text-sm text-base-content/70 bg-info bg-opacity-10 p-3 rounded-lg">
        <strong>Note:</strong> These are project-level domain settings that will be used as defaults for new deployments. 
        Changes here will affect conflict detection and ensure domains don't conflict with other projects.
      </div>

      {error() && (
        <div class="alert alert-error">
          <span>{error()}</span>
        </div>
      )}

      {/* Add New Domain */}
      <div class="card bg-base-50">
        <div class="card-body">
          <h4 class="font-medium mb-3">Add New Domain</h4>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
            <div class="form-control">
              <label class="label">
                <span class="label-text">Domain</span>
              </label>
              <input
                type="text"
                placeholder="example.com"
                value={newDomain()}
                onInput={(e) => setNewDomain(e.target.value)}
                class="input input-bordered input-sm"
                disabled={loading()}
              />
            </div>
            <div class="form-control">
              <label class="label">
                <span class="label-text">Port</span>
              </label>
              <input
                type="number"
                placeholder="8080"
                value={newPort()}
                onInput={(e) => setNewPort(e.target.value)}
                class="input input-bordered input-sm"
                disabled={loading()}
              />
            </div>
            <div class="form-control">
              <label class="label">
                <span class="label-text">Host</span>
              </label>
              <input
                type="text"
                placeholder="localhost"
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
                  Adding...
                </>
              ) : (
                'Add Domain'
              )}
            </button>
          </div>
        </div>
      </div>

      {/* Current Domain Configuration */}
      <div class="card bg-base-100">
        <div class="card-body">
          <h4 class="font-medium mb-3">Current Domain Configuration</h4>
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
                              Proxy to: {domain.proxy_target} (Port: {domain.port}, Host: {domain.host})
                            </div>
                          )}
                        </div>
                      </div>
                      <div class="flex items-center gap-2">
                        {domain.configured ? (
                          <div class="badge badge-success badge-sm">Configured</div>
                        ) : (
                          <div class="badge badge-warning badge-sm">Not Configured</div>
                        )}
                        <button
                          class="btn btn-error btn-xs"
                          onClick={() => removeDomain(domain.domain)}
                          disabled={loading()}
                        >
                          Remove
                        </button>
                      </div>
                    </div>
                  )}
                </For>
              </div>
            ) : (
              <div class="text-center py-8 text-base-content/60">
                No domains configured for this project
              </div>
            )
          ) : (
            <div class="flex items-center justify-center py-8">
              <span class="loading loading-spinner loading-lg"></span>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default ProjectDomainManagement