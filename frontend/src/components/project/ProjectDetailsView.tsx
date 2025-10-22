import type { Component } from 'solid-js'
import type { Project } from '../../types/project'
import ProjectOverview from './ProjectOverview'

interface ProjectDetailsViewProps {
  project: Project
  onBack: () => void
}

const ProjectDetailsView: Component<ProjectDetailsViewProps> = (props) => {

  return (
    <div class="card bg-base-100 shadow mb-6">
      <div class="card-body">
        <div class="flex items-start justify-between">
          <div>
            <h2 class="text-2xl font-bold">{props.project.name}</h2>
            <div class="text-sm text-base-content/70">
              <a class="link" href={props.project.gitRepository} target="_blank" rel="noreferrer">
                {props.project.gitRepository}
              </a>
            </div>
          </div>
        
        </div>

        {/* Overview Content */}
        <div class="mt-4">
          <ProjectOverview 
            project={props.project}
            showProjectInfo={false}
          />
        </div>
      </div>
    </div>
  )
}

export default ProjectDetailsView