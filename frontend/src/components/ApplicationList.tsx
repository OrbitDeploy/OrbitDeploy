import { createSignal, Show, For, createEffect } from 'solid-js'
import type { Component } from 'solid-js'
import { useNavigate } from '@solidjs/router'
import type { Application } from '../types/project'
import { useApiQuery } from '../api/apiHooksW.ts'
import { listAppsEndpoint, listAppsByNameEndpoint } from '../api/endpoints/projects'
import { useI18n } from '../i18n'

interface ApplicationListProps {
  projectUid: string
  projectName?: string  // Add optional project name
  onCreateApplication: () => void
}

const ApplicationList: Component<ApplicationListProps> = (props) => {
  const navigate = useNavigate()
  const { t } = useI18n()

  // Edit modal state
  const [isEditModalOpen, setIsEditModalOpen] = createSignal(false)
  const [editingApplication, setEditingApplication] = createSignal<Application | null>(null)

  // Query applications for this project
  const applicationsQuery = useApiQuery<Application[]>(
    () => ['projects', props.projectUid || props.projectName, 'applications'],
    () => props.projectUid 
      ? listAppsEndpoint(props.projectUid).url
      : listAppsByNameEndpoint(props.projectName!).url,
    {
      enabled: () => !!props.projectUid || !!props.projectName,
    }
  )

  const applications = () => {
    const apps = (applicationsQuery.data || []).map(app => ({
      uid: app.uid,
      projectUid: app.projectUid,
      name: app.name,
      description: app.description,
      activeReleaseUid: app.activeReleaseUid,
      repoUrl: app.repoUrl,
      buildDir: app.buildDir,
      buildType: app.buildType,
      targetPort: app.targetPort,
      status: app.status,
      volumes: app.volumes,
      execCommand: app.execCommand,
      autoUpdatePolicy: app.autoUpdatePolicy,
      branch: app.branch,
      createdAt: app.createdAt,
      updatedAt: app.updatedAt,
    }))
    return apps
  }

  const handleViewApp = (app: Application) => {
    navigate(`/projects/${props.projectUid}/apps/${app.uid}`)
  }

  const handleEditApp = (app: Application) => {
    setEditingApplication(app)
    setIsEditModalOpen(true)
  }

  const handleEditSuccess = (message: string) => {
    // Refresh the applications list
    applicationsQuery.refetch()
    // You might want to show a toast notification here
    console.log(message)
  }

  const handleEditError = (message: string) => {
    // You might want to show an error notification here
    console.error(message)
  }

  const handleCloseEditModal = () => {
    setIsEditModalOpen(false)
    setEditingApplication(null)
  }

  return (
    <div class="card bg-base-100 shadow-xl">
      <div class="card-body">
        <div class="flex justify-between items-center mb-4">
          <h3 class="card-title">{t('application_list.title')}</h3>
          <button 
            class="btn btn-primary btn-sm"
            onClick={props.onCreateApplication}
          >
            {t('application_list.add_new_button')}
          </button>
        </div>

        <Show 
          when={!applicationsQuery.isPending} 
          fallback={
            <div class="flex justify-center py-8">
              <span class="loading loading-spinner loading-lg"></span>
            </div>
          }
        >
          <Show 
            when={applications().length > 0}
            fallback={
              <div class="text-center py-8 text-base-content/70">
                <p>{t('application_list.empty_state_title')}</p>
                <p class="text-sm mt-2">{t('application_list.empty_state_description')}</p>
              </div>
            }
          >
            <div class="overflow-x-auto">
              <table class="table table-zebra">
                <thead>
                  <tr>
                    <th>{t('application_list.table_header_name')}</th>
                    <th>{t('application_list.table_header_description')}</th>
                    <th>{t('application_list.table_header_repo')}</th>
                    <th>{t('application_list.table_header_port')}</th>
                    <th>{t('application_list.table_header_status')}</th>
                    <th>{t('application_list.table_header_created')}</th>
                    {/* <th>{t('application_list.table_header_actions')}</th> */}
                  </tr>
                </thead>
                <tbody>
                  <For each={applications()}>
                    {(app) => (
                      <tr class="cursor-pointer" onClick={() => handleViewApp(app)}>
                        <td>
                          <div class="font-medium">{app.name}</div>
                        </td>
                        <td>
                          <div class="text-sm text-base-content/70 truncate max-w-xs">
                            {app.description || '-'}
                          </div>
                        </td>
                        <td>
                          <div class="truncate max-w-xs">
                            {app.repoUrl || '-'}
                          </div>
                        </td>
                        <td>
                          <div class="badge badge-outline">{app.targetPort}</div>
                        </td>
                        <td>
                          <div class={`badge ${
                            app.status === 'running' ? 'badge-success' :
                            app.status === 'stopped' ? 'badge-error' :
                            'badge-warning'
                          }`}>
                            {app.status === 'running' ? t('application_list.status_running') :
                             app.status === 'stopped' ? t('application_list.status_stopped') :
                             t('application_list.status_unknown')}
                          </div>
                        </td>
                        <td>
                          <div class="text-sm">
                            {app.createdAt ? new Date(app.createdAt).toLocaleDateString() : '-'}
                          </div>
                        </td>
                        <td>
                          {/* <div class="flex gap-2">
                            <button 
                              class="btn btn-sm btn-outline"
                              onClick={() => handleViewApp(app)}
                            >
                              {t('common.view')}
                            </button>
                            <button 
                              class="btn btn-sm btn-outline"
                              onClick={() => handleEditApp(app)}
                            >
                              {t('common.edit')}
                            </button>
                          </div> */}
                        </td>
                      </tr>
                    )}
                  </For>
                </tbody>
              </table>
            </div>
          </Show>
        </Show>

        {/* <EditApplicationModal
          isOpen={isEditModalOpen()}
          projectUid={props.projectUid}
          application={editingApplication()}
          onClose={handleCloseEditModal}
          onSuccess={handleEditSuccess}
          onError={handleEditError}
          onRefresh={() => applicationsQuery.refetch()}
        /> */}

        <Show when={applicationsQuery.error}>
          <div class="alert alert-error mt-4">
            <span>{t('application_list.load_error_prefix')} {applicationsQuery.error.message}</span>
          </div>
        </Show>
      </div>
    </div>
  )
}

export default ApplicationList