import { createSignal } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n } from '../../i18n'
import type { Project } from '../../types/project'
import { useApiMutation } from '../../lib/apiHooks'
import { getProjectsApiUrl } from '../../api/config'

interface CreateProjectModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: (message: string) => void
  onError: (message: string) => void
  onRefresh: () => void
}

const CreateProjectModal: Component<CreateProjectModalProps> = (props) => {
  const { t } = useI18n()

  // Create project form
  const [newProject, setNewProject] = createSignal<Partial<Project>>({
    name: '',
    description: '',
  })

  // Create project via real API; on failure show error; no mock creation
  const createMutation = useApiMutation<unknown, { payload: any }>(
    getProjectsApiUrl('create'),
    {
      method: 'POST',
      onSuccess: () => {
        props.onClose()
        resetForm()
        props.onSuccess(t('projects.messages.create_success') || 'Created')
        props.onRefresh()
      },
      onError: (err: Error) => {
        props.onError(err.message)
      }
    }
  )

  function createProject() {
    const form = newProject()
 
    // Validate project name (no uppercase)
    if (form.name?.trim() !== form.name?.trim().toLowerCase()) {
      props.onError('Project name cannot contain uppercase letters')
      return
    }
    
    // Create payload matching the API handler format
    const payload: any = {
      name: form.name!.trim(),
      description: form.description || '',
    }

    createMutation.mutate(payload)
  }

  const resetForm = () => {
    setNewProject({
      name: '',
      description: '',
    })
  }

  return (
    <div class={`modal ${props.isOpen ? 'modal-open' : ''}`}>
      <div class="modal-box max-w-2xl">
        <h3 class="font-bold text-lg mb-4">{t('projects.modals.create_title')}</h3>
        <div class="grid grid-cols-1 gap-3">
          <div>
            <label class="label"><span class="label-text">{t('projects.form.name')}</span></label>
            <input class="input input-bordered w-full" value={newProject().name || ''} onInput={(e) => setNewProject(p => ({ ...p, name: e.currentTarget.value }))} />
          </div>
          <div>
            <label class="label"><span class="label-text">{t('projects.form.description')}</span></label>
            <textarea class="textarea textarea-bordered w-full" value={newProject().description || ''} onInput={(e) => setNewProject(p => ({ ...p, description: e.currentTarget.value }))} />
          </div>
        </div>
        <div class="modal-action">
          <button class="btn btn-primary" disabled={createMutation.isPending} onClick={() => void createProject()}>{t('common.save')}</button>
          <button class="btn" onClick={props.onClose}>{t('common.cancel')}</button>
        </div>
      </div>
    </div>
  )
}

export default CreateProjectModal