import { createSignal, Show, For, Component, createEffect } from 'solid-js'
import { useQueryClient } from '@tanstack/solid-query'
import { toast } from 'solid-toast'
import { useI18n } from '../../i18n'
import { useApiQuery, useApiMutation } from '../../api/apiHooksW.ts'
import { 
  listRoutingsEndpoint, 
  createRoutingEndpoint, 
  updateRoutingEndpoint, 
  deleteRoutingEndpoint,
  getLatestApplicationReleaseEndpoint 
} from '../../api/endpoints'
import type { RoutingResponse, RoutingRequest, RoutingsResponse } from '../../types/routing'
import type { Release } from '../../types/deployment'

interface ApplicationDomainTabProps {
  applicationUid: string
}

const ApplicationDomainTab: Component<ApplicationDomainTabProps> = (props) => {
  const { t } = useI18n()
  const queryClient = useQueryClient()

  // State for modal and form
  const [showAddModal, setShowAddModal] = createSignal(false)
  const [editingRouting, setEditingRouting] = createSignal<RoutingResponse | null>(null)
  const [domainName, setDomainName] = createSignal('')
  const [hostPort, setHostPort] = createSignal('8080')
  const [isActive, setIsActive] = createSignal(true)
  const [error, setError] = createSignal('')

  // Query for listing routings
  const routingsQuery = useApiQuery<RoutingsResponse>(
    () => ['routings', props.applicationUid],
    () => listRoutingsEndpoint(props.applicationUid).url
  )

  // Query for getting the latest release to get the default port
  const latestReleaseQuery = useApiQuery<Release | null>(
    () => ['latestRelease', props.applicationUid],
    () => getLatestApplicationReleaseEndpoint(props.applicationUid).url
  )

  // Get the default port from the latest release
  const getDefaultPort = () => {
    const release = latestReleaseQuery.data
    if (release && release.systemPort && release.systemPort > 0) {
      return release.systemPort.toString()
    }
    return '8080' // Default fallback port
  }

  const refreshRoutings = async () => {
    await queryClient.invalidateQueries({ queryKey: ['routings', props.applicationUid] })
  }

  // Update default port when latest release data is loaded
  createEffect(() => {
    if (latestReleaseQuery.data) {
      const newPort = getDefaultPort()
      // Update if modal is closed, or if modal is open and port is still the fallback
      if (!showAddModal() || hostPort() === '8080') {
        setHostPort(newPort)
      }
    }
  })

  // Mutations for CRUD operations
  const createMutation = useApiMutation<RoutingResponse, RoutingRequest>(
    createRoutingEndpoint(props.applicationUid),
    {
      onSuccess: () => {
        setShowAddModal(false)
        resetForm()
        toast.success(t('domain_tab.add_success_toast'))
        void refreshRoutings()
      },
      onError: (error: Error) => {
        setError(error.message || t('domain_tab.add_error_toast'))
      }
    }
  )

  const updateMutation = useApiMutation<RoutingResponse, RoutingRequest>(
    (data) => {
      const routing = editingRouting()
      if (!routing) throw new Error('No routing being edited')
      return updateRoutingEndpoint(routing.uid)
    },
    {
      onSuccess: () => {
        setEditingRouting(null)
        resetForm()
        toast.success(t('domain_tab.update_success_toast'))
        void refreshRoutings()
      },
      onError: (error: Error) => {
        setError(error.message || t('domain_tab.update_error_toast'))
      }
    }
  )

  const deleteMutation = useApiMutation<unknown, { uid: string }>(
    (variables: { uid: string }) => deleteRoutingEndpoint(variables.uid),
    {
      onSuccess: () => {
        toast.success(t('domain_tab.delete_success_toast'))
        void refreshRoutings()
      },
      onError: (error: Error) => {
        toast.error(error.message || t('domain_tab.delete_error_toast'))
      }
    }
  )

  // Update routings accessor to extract from response
  const routings = () => routingsQuery.data?.routings || []
  const isLoading = () => routingsQuery.isPending
  const queryError = () => routingsQuery.error
  const isMutating = () => createMutation.isPending || updateMutation.isPending || deleteMutation.isPending

  const resetForm = () => {
    setDomainName('')
    setHostPort(getDefaultPort())
    setIsActive(true)
    setError('')
  }

  const openAddModal = () => {
    resetForm()
    setEditingRouting(null)
    setShowAddModal(true)
  }

  const openEditModal = (routing: RoutingResponse) => {
    setDomainName(routing.domainName)
    setHostPort(routing.hostPort.toString())
    setIsActive(routing.isActive)
    setEditingRouting(routing)
    setShowAddModal(true)
    setError('')
  }

  const handleSubmit = () => {
    if (!domainName().trim()) {
      setError(t('domain_tab.error_domain_required'))
      return
    }

    const port = parseInt(hostPort())
    if (isNaN(port) || port <= 0 || port > 65535) {
      setError(t('domain_tab.error_invalid_port'))
      return
    }

    setError('')

    if (editingRouting()) {
      // Update existing routing
      updateMutation.mutate({
        domainName: domainName().trim(),
        hostPort: port,
        isActive: isActive()
      })
    } else {
      // Create new routing
      createMutation.mutate({
        domainName: domainName().trim(),
        hostPort: port,
        isActive: isActive()
      })
    }
  }

  const handleDelete = (routing: RoutingResponse) => {
    if (confirm(t('domain_tab.delete_confirm', { domainName: routing.domainName }))) {
      deleteMutation.mutate({ uid: routing.uid })
    }
  }

  const formatDate = (dateStr: string) => {
    try {
      return new Date(dateStr).toLocaleString()
    } catch {
      return dateStr
    }
  }

  return (
    <div class="space-y-4">
      <div class="flex justify-between items-center">
        <h3 class="text-lg font-semibold">{t('domain_tab.title')}</h3>
        <button 
          class="btn btn-primary btn-sm"
          onClick={openAddModal}
          disabled={isMutating()}
        >
          {t('domain_tab.add_domain_button')}
        </button>
      </div>

      {/* Loading state */}
      <Show when={isLoading()}>
        <div class="flex justify-center p-6">
          <span class="loading loading-spinner loading-lg"></span>
        </div>
      </Show>

      {/* Error state */}
      <Show when={queryError()}>
        {(err) => (
          <div class="alert alert-error">
            <span>{t('domain_tab.load_error_prefix')} {err instanceof Error ? err.message : t('domain_tab.unknown_error')}</span>
          </div>
        )}
      </Show>

      {/* Routings table */}
      <Show when={!queryError()}>
        <div class="card bg-base-100 shadow-xl">
          <div class="card-body">
            <Show when={routings().length === 0}
                  fallback={
                    <div class="overflow-x-auto">
                      <table class="table table-sm">
                        <thead>
                          <tr>
                            <th>{t('domain_tab.table_header_domain')}</th>
                            <th>{t('domain_tab.table_header_port')}</th>
                            <th>{t('domain_tab.table_header_status')}</th>
                            <th>{t('domain_tab.table_header_created')}</th>
                            <th>{t('domain_tab.table_header_actions')}</th>
                          </tr>
                        </thead>
                        <tbody>
                          <For each={routings()}>{(routing) => (
                            <tr>
                              <td class="font-mono text-sm">
                                {/* Check if it's a valid domain for linking */}
                                {/^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(routing.domainName) ? (
                                  <a
                                    href={`https://${routing.domainName}`}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    class="link link-primary"
                                    title={t('domain_tab.open_domain_title', { domainName: routing.domainName })}
                                  >
                                    {routing.domainName}
                                  </a>
                                ) : (
                                  <span>{routing.domainName}</span>
                                )}
                              </td>
                              <td class="text-sm">{routing.hostPort}</td>
                              <td>
                                <span class={`badge badge-sm ${                                  routing.isActive ? 'badge-success' : 'badge-outline'
                                }`}>
                                  {routing.isActive ? t('domain_tab.status_active') : t('domain_tab.status_inactive')}
                                </span>
                              </td>
                              <td class="text-sm">{formatDate(routing.createdAt)}</td>
                              <td>
                                <div class="flex items-center gap-2">
                                  <button
                                    class="btn btn-ghost btn-xs"
                                    onClick={() => openEditModal(routing)}
                                    disabled={isMutating()}
                                  >
                                    {t('common.edit')}
                                  </button>
                                  <button
                                    class="btn btn-ghost btn-xs text-error"
                                    onClick={() => handleDelete(routing)}
                                    disabled={isMutating()}
                                  >
                                    {t('common.delete')}
                                  </button>
                                </div>
                              </td>
                            </tr>
                          )}</For>
                        </tbody>
                      </table>
                    </div>
                  }
            >
              <div class="text-center py-8 text-base-content/70">
                {t('domain_tab.no_domains_message')}
              </div>
            </Show>
          </div>
        </div>
      </Show>

      {/* Add/Edit Modal */}
      <div class={`modal ${showAddModal() ? 'modal-open' : ''}`}>
        <div class="modal-box">
          <div class="flex items-center justify-between mb-4">
            <h3 class="font-bold text-lg">
              {editingRouting() ? t('domain_tab.modal_title_edit') : t('domain_tab.modal_title_add')}
            </h3>
            <button 
              class="btn btn-sm btn-circle btn-ghost"
              onClick={() => setShowAddModal(false)}
            >
              ✕
            </button>
          </div>

          {error() && (
            <div class="alert alert-error mb-4">
              <span>{error()}</span>
            </div>
          )}

          <div class="space-y-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">{t('domain_tab.form_label_domain')}</span>
              </label>
              <input
                type="text"
                placeholder={t('domain_tab.form_placeholder_domain')}
                value={domainName()}
                onInput={(e) => setDomainName(e.currentTarget.value)}
                class="input input-bordered"
                disabled={isMutating()}
              />
            </div>

            <div class="form-control">
              <label class="label">
                <span class="label-text">{t('domain_tab.form_label_port')}</span>
              </label>
              <input
                type="number"
                placeholder={t('domain_tab.form_placeholder_port')}
                value={hostPort()}
                onInput={(e) => setHostPort(e.currentTarget.value)}
                class="input input-bordered"
                min="1"
                max="65535"
                disabled={isMutating()}
              />
            </div>

            <div class="form-control">
              <label class="cursor-pointer label">
                <span class="label-text">{t('domain_tab.form_label_enable')}</span>
                <input
                  type="checkbox"
                  checked={isActive()}
                  onChange={(e) => setIsActive(e.currentTarget.checked)}
                  class="checkbox checkbox-primary"
                  disabled={isMutating()}
                />
              </label>
            </div>
          </div>

          <div class="modal-action">
            <button
              class="btn"
              onClick={() => setShowAddModal(false)}
              disabled={isMutating()}
            >
              {t('common.cancel')}
            </button>
            <button
              class="btn btn-primary"
              onClick={handleSubmit}
              disabled={isMutating() || !domainName().trim()}
            >
              {isMutating() && <span class="loading loading-spinner loading-sm"></span>}
              {isMutating() ? t('domain_tab.saving_button') : (editingRouting() ? t('domain_tab.update_button') : t('domain_tab.add_button'))}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default ApplicationDomainTab