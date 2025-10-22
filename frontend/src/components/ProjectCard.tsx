import type { Component } from 'solid-js'
import { Show, For } from 'solid-js'
import { useNavigate } from '@solidjs/router'
import { useI18n } from '../i18n'
import type { Project, Application } from '../types/project'
import { useApiQuery } from '../lib/apiHooks'
import { listAppsEndpoint } from '../api/endpoints/projects'

interface ProjectCardProps {
  project: Project
  onSelect: (project: Project) => void
  onDelete: (project: Project) => void
  showApplications?: boolean // New prop to control whether to show applications
}

const ProjectCard: Component<ProjectCardProps> = (props) => {
  const { t } = useI18n()
  const navigate = useNavigate()

  console.log('ProjectCard showApplications:', props.showApplications) // Debug: check prop in ProjectCard

  // Query applications for this project if showApplications is true
  const applicationsQuery = useApiQuery<Application[]>(
    () => ['projects', props.project.uid, 'applications'],
    () => listAppsEndpoint(props.project.uid).url,
    {
      enabled: () => {
        const enabled = !!props.showApplications
        console.log('Applications query enabled:', enabled, 'for project:', props.project.uid) // Debug: log enabled state
        return enabled
      },
    }
  )

  console.log('Applications query status:', { isPending: applicationsQuery.isPending, isError: applicationsQuery.isError, data: applicationsQuery.data }) // Debug: log query status

  const applications = () => {
    const apps = applicationsQuery.data || []
    console.log('Applications data:', apps) // Debug: log applications array
    return apps
  }

  const handleAppClick = (app: Application, e: Event) => {
    e.stopPropagation()
    navigate(`/projects/${props.project.name}/apps/${app.name}`)
  }

  const handleCreateApp = (e: Event) => {
    e.stopPropagation()
    // For now, navigate to project detail page where they can create apps
    props.onSelect(props.project)
  }

  return (
    <div class={`card bg-base-200 hover:bg-base-300 transition cursor-pointer ${props.showApplications ? 'w-full max-w-lg' : ''}`} 
         onClick={() => props.onSelect(props.project)}>
      <div class="card-body">
        <div class="flex items-start justify-between">
          <div class="flex-1">
            <div class="text-lg font-semibold">{props.project.name}</div>
            <div class="text-xs font-mono text-base-content/70 truncate max-w-[240px]">
              {props.project.gitRepository}
            </div>
          </div>
        
        </div>

        {/* Applications Section */}
        <Show when={props.showApplications}>
          <div class="mt-4 border-t border-base-300 pt-4">
            <div class="flex justify-between items-center mb-2">
              <h4 class="text-sm font-medium text-base-content/80">应用列表</h4>
              <Show when={!applicationsQuery.isPending}>
                <span class="text-xs text-base-content/60">
                  {applications().length} 个应用
                </span>
              </Show>
            </div>
            
            <Show 
              when={!applicationsQuery.isPending}
              fallback={
                <div class="flex justify-center py-2">
                  <span class="loading loading-spinner loading-sm"></span>
                </div>
              }
            >
              <Show 
                when={applications().length > 0}
                fallback={
                  <div class="text-center py-4 text-sm text-base-content/60">
                    <p>暂无应用</p>
                    <button 
                      class="btn btn-xs btn-outline mt-2"
                      onClick={handleCreateApp}
                    >
                      添加应用
                    </button>
                  </div>
                }
              >
                <div class="space-y-2">
                  <For each={applications()}>
                    {(app) => {
                      console.log('Rendering app:', app) // Debug: log each app being rendered
                      return (
                        <div 
                          class="flex items-center justify-between p-2 bg-base-100 rounded-md hover:bg-base-300 transition cursor-pointer"
                          onClick={(e) => handleAppClick(app, e)}
                        >
                          <div class="flex-1">
                            <div class="font-medium text-sm">{app.name}</div>
                            <div class="text-xs text-base-content/70 truncate">
                              {app.description || '无描述'}
                            </div>
                          </div>
                          <div class="flex items-center gap-2">
                            <span class="text-xs badge badge-outline">
                              :{app.targetPort}
                            </span>
                            <span class={`badge badge-xs ${
                              app.status === 'running' ? 'badge-success' :
                              app.status === 'stopped' ? 'badge-error' :
                              'badge-warning'
                            }`}>
                              {app.status}
                            </span>
                          </div>
                        </div>
                      )
                    }}
                  </For>
                  
                  {/* Add App Button */}
                  <button 
                    class="btn btn-xs btn-outline w-full"
                    onClick={handleCreateApp}
                  >
                    + 添加应用
                  </button>
                </div>
              </Show>
            </Show>
          </div>
        </Show>

        <div class="mt-2 text-xs text-base-content/70">
          {props.project.createdAt ? new Date(props.project.createdAt).toLocaleString() : '-'}
        </div>
      </div>
    </div>
  )
}

export default ProjectCard