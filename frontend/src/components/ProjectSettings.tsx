import type { Component } from 'solid-js'
import { useI18n } from '../i18n'
import type { Project } from '../types/project'

interface ProjectSettingsProps {
  project: Project
}

const ProjectSettings: Component<ProjectSettingsProps> = (props) => {
  const { t } = useI18n()

  return (
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div>
        <label class="label"><span class="label-text">{t('projects.form.dockerfile')}</span></label>
        <input 
          class="input input-bordered w-full" 
          value={props.project.dockerfile || 'Dockerfile'} 
          disabled 
        />
      </div>
      <div>
        <label class="label"><span class="label-text">{t('projects.form.context')}</span></label>
        <input 
          class="input input-bordered w-full" 
          value={props.project.context_path || '.'} 
          disabled 
        />
      </div>
      <div class="md:col-span-2">
        <label class="label"><span class="label-text">{t('projects.form.description')}</span></label>
        <textarea 
          class="textarea textarea-bordered w-full" 
          value={props.project.description || ''} 
          disabled 
        />
      </div>
    </div>
  )
}

export default ProjectSettings