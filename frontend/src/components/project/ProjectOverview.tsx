import { Show } from 'solid-js'
import type { Component } from 'solid-js'
import type { Project } from '../../types/project'

interface ProjectOverviewProps {
  project: Project
  showProjectInfo?: boolean
}

const ProjectOverview: Component<ProjectOverviewProps> = (props) => {
  const showProjectInfo = () => props.showProjectInfo !== false

  return (
    <div>
      {/* Project Information */}
      <Show when={showProjectInfo()}>
        <div class="card bg-base-200">
          <div class="card-body p-4">
            <h4 class="font-semibold mb-2">Project Information</h4>
            <div class="space-y-2 text-sm">
              <div><strong>Repository:</strong> {props.project.GitRepository}</div>
              <div><strong>Branch:</strong> {props.project.GitBranch || 'main'}</div>
              <div><strong>Dockerfile:</strong> {props.project.Dockerfile || 'Dockerfile'}</div>
              <div><strong>Context Path:</strong> {props.project.ContextPath || '.'}</div>
              {props.project.PublishPort && (
                <div><strong>Default Port:</strong> {props.project.PublishPort}</div>
              )}
              {props.project.Description && (
                <div><strong>Description:</strong> {props.project.Description}</div>
              )}
            </div>
          </div>
        </div>
      </Show>
    </div>
  )
}

export default ProjectOverview