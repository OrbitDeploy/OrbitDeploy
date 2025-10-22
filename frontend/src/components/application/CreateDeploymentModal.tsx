import { Component, createSignal, Show, For } from 'solid-js'
import type { Application } from '../../types/project'
import { useApiQuery, useApiMutation } from '../../api/apiHooksW.ts'
import { 
  getApplicationReleasesEndpoint, 
  createApplicationDeploymentEndpoint 
} from '../../api/endpoints'
import { useQueryClient } from '@tanstack/solid-query'
import { toast } from 'solid-toast'
import { Release } from '../../types/deployment'
import { useI18n } from '../../i18n'


interface CreateDeploymentModalProps {
  isOpen: boolean
  onClose: () => void
  application: Application | null
}

const CreateDeploymentModal: Component<CreateDeploymentModalProps> = (props) => {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const [selectedOption, setSelectedOption] = createSignal<string>('')
  const [isRebuild, setIsRebuild] = createSignal(true)

  // Query to get available releases for this application
  const releasesQuery = useApiQuery<Release[]>(
    () => ['applications', props.application?.uid, 'releases'],
    () => {
      if (!props.application?.uid) return null
      return getApplicationReleasesEndpoint(props.application.uid).url
    },
    // Only fetch releases if the modal is open AND the user has unchecked "Rebuild"
    { enabled: () => !!(props.isOpen && props.application?.uid && !isRebuild()) }
  )

  // Mutation to create deployment
  const createDeploymentMutation = useApiMutation<any, {
    releaseId: string | null
  }>(
    () => {
      if (!props.application?.uid) throw new Error('No application ID')
      return createApplicationDeploymentEndpoint(props.application.uid)
    },
    {
      onSuccess: () => {
        toast.success(t('create_deployment_modal.success_toast'))
        queryClient.invalidateQueries({ queryKey: ['applications', props.application?.uid, 'deployments'] })
        handleClose()
      },
      onError: (error: Error) => {
        toast.error(`${t('create_deployment_modal.error_toast_prefix')} ${error.message}`)
      }
    }
  )

  const releases = () => releasesQuery.data || []

  const handleSubmit = () => {
    const releaseId = isRebuild() ? null : selectedOption()
    if (!isRebuild() && !releaseId) {
      toast.error(t('create_deployment_modal.error_no_release_selected'))
      return
    }

    createDeploymentMutation.mutate({
      releaseId
    })
  }

  const handleClose = () => {
    setSelectedOption('')
    setIsRebuild(true)
    props.onClose()
  }

  return (
    <Show when={props.isOpen}>
      <div class="modal modal-open">
        <div class="modal-box">
          <h3 class="font-bold text-lg mb-4">{t('create_deployment_modal.title')}</h3>

          <div class="space-y-4">
            {/* Application Info */}
            <div class="bg-base-200 p-3 rounded">
              <p class="text-sm text-base-content/70">{t('create_deployment_modal.app')}: {props.application?.name}</p>
              <p class="text-sm text-base-content/70">{t('create_deployment_modal.description')}: {props.application?.description}</p>
            </div>

            {/* Rebuild Checkbox */}
            <div class="form-control">
              <label class="label cursor-pointer">
                <span class="label-text">{t('create_deployment_modal.rebuild_option')}</span>
                <input 
                  type="checkbox" 
                  class="checkbox" 
                  checked={isRebuild()} 
                  onChange={(e) => setIsRebuild(e.target.checked)} 
                />
              </label>
            </div>

            {/* Release Selection */}
            <div class="form-control">
              <label class="label">
                <span class="label-text">{t('create_deployment_modal.select_release_label')}</span>
              </label>
              <Show 
                when={!isRebuild()}
                fallback={
                  <div class="alert alert-info">
                    <span>{t('create_deployment_modal.info_rebuild_selected')}</span>
                  </div>
                }
              >
                <Show
                  when={!releasesQuery.isPending}
                  fallback={
                    <div class="flex items-center gap-2">
                      <span class="loading loading-spinner loading-sm"></span>
                      <span class="text-sm">{t('create_deployment_modal.loading_releases')}</span>
                    </div>
                  }
                >
                  <Show
                    when={releases().length > 0}
                    fallback={
                      <div class="alert alert-warning">
                        <span>{t('create_deployment_modal.warn_no_releases')}</span>
                      </div>
                    }
                  >
                    <select 
                      class="select select-bordered w-full"
                      value={selectedOption()}
                      onChange={(e) => setSelectedOption(e.target.value)}
                      disabled={isRebuild()}
                    >
                      <option value="">{t('create_deployment_modal.select_placeholder')}</option>
                      <For each={releases()}>
                        {(release) => (
                          <option value={release.uid}>
                            {`${release.imageName} (${t('create_deployment_modal.status_label')}: ${release.status}) - ${release.createdAt}`}
                          </option>
                        )}
                      </For>
                    </select>
                  </Show>
                </Show>
              </Show>
            </div>
          </div>

          <div class="modal-action">
            <button 
              class="btn btn-primary"
              onClick={handleSubmit}
              disabled={createDeploymentMutation.isPending || (!isRebuild() && !selectedOption())}
            >
              {createDeploymentMutation.isPending ? (
                <>
                  <span class="loading loading-spinner loading-sm"></span>
                  {t('create_deployment_modal.creating_button')}
                </> 
              ) : (
                t('create_deployment_modal.create_button')
              )}
            </button>
            <button class="btn" onClick={handleClose}>
              {t('create_deployment_modal.cancel_button')}
            </button>
          </div>
        </div>
      </div>
    </Show>
  )
}

export default CreateDeploymentModal