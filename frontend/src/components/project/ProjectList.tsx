import { Show, For } from 'solid-js'
import type { Component, Resource } from 'solid-js'
import { useI18n } from '../../i18n'
import type { Project } from '../../types/project'
import ProjectCard from '../ProjectCard'

interface ProjectListProps {
  projectsData: Resource<Project[] | undefined>
  onSelectProject: (project: Project) => void
  onDeleteProject: (project: Project) => void
  showApplications?: boolean
}

const ProjectList: Component<ProjectListProps> = (props) => {
  const { t } = useI18n()

  console.log('Projects data loading:', props.projectsData.loading) // Debug: log loading state
  console.log('Projects data:', props.projectsData()) // Debug: log projects array
  console.log('showApplications prop:', props.showApplications) // Debug: check showApplications value

  return (
    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <h2 class="card-title">{t('projects.list_title')}</h2>
        <Show when={!props.projectsData.loading && props.projectsData()?.length}
              fallback={
                <Show when={props.projectsData.loading}
                      fallback={<div class="text-center py-12 text-base-content/70">{t('projects.empty_state')}</div>}>
                  <div class="text-center py-12">
                    <span class="loading loading-spinner loading-lg"></span>
                    <p class="mt-4 text-base-content/70">{t('common.loading')}</p>
                  </div>
                </Show>
              }>
          <div class={`grid gap-4 ${props.showApplications ? 'grid-cols-1 lg:grid-cols-2' : 'grid-cols-1 sm:grid-cols-2 xl:grid-cols-3'}`}>
            <For each={props.projectsData() || []}>{(project) => {
              console.log('Rendering project:', project) // Debug: log each project being rendered
              console.log('Passing showApplications to ProjectCard:', props.showApplications) // Debug: confirm prop passed
              return (
                <ProjectCard 
                  project={project}
                  onSelect={props.onSelectProject}
                  onDelete={props.onDeleteProject}
                  showApplications={props.showApplications}
                />
              )
            }}</For>
          </div>
        </Show>
      </div>
    </div>
  )
}

export default ProjectList