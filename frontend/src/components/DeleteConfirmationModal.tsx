import { createSignal, Show } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n } from '../i18n'
import type { Project } from '../types/project'

interface DeleteConfirmationModalProps {
  isOpen: boolean
  project: Project | null
  onClose: () => void
  onSuccess: (message: string) => void
  onError: (message: string) => void
  onRefresh: () => void
}

const DeleteConfirmationModal: Component<DeleteConfirmationModalProps> = (props) => {
  const { t } = useI18n()
  const [deleteConfirmName, setDeleteConfirmName] = createSignal('')
  const [error, setError] = createSignal('')

  async function confirmDelete() {
    const project = props.project
    if (!project) return
    
    setError('')
    
    // Verify project name matches
    if (deleteConfirmName().trim() !== project.Name) {
      setError('Project name confirmation does not match')
      return
    }
    
    try {
      const res = await fetch(`/api/projects/${project.uid}`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          project_name: project.Name
        })
      })
      const json = await res.json()
      if (!res.ok || !json?.success) {
        const msg = (json && json.message) ? json.message : 'Delete project failed'
        throw new Error(msg)
      }
      
      props.onClose()
      setDeleteConfirmName('')
      setError('')
      props.onSuccess('Project deleted successfully')
      props.onRefresh()
    } catch (e) {
      setError((e as Error).message)
    }
  }

  function handleClose() {
    setDeleteConfirmName('')
    setError('')
    props.onClose()
  }

  

  return (
    <div class={`modal ${props.isOpen && props.project ? 'modal-open' : ''}`}>
      <div class="modal-box">
        <h3 class="font-bold text-lg mb-4 text-error">Delete Project</h3>
        <div class="space-y-4">
          <div class="alert alert-warning">
            <div>
              <strong>Warning:</strong> This action cannot be undone. This will:
              <ul class="list-disc list-inside mt-2 space-y-1">
                <li>Stop and delete all associated containers</li>
                <li>Remove all domain associations</li>
                <li>Delete all images built through this project</li>
                <li>Delete all build and deployment history</li>
              </ul>
            </div>
          </div>
          <div>
            <p class="mb-2">
              Please type <strong>{props.project?.Name || ''}</strong> to confirm deletion:
            </p>
            <input 
              class="input input-bordered w-full" 
              value={deleteConfirmName()} 
              onInput={(e) => setDeleteConfirmName(e.currentTarget.value)}
              placeholder="Enter project name"
            />
          </div>
          <Show when={error()}>
            <div class="alert alert-error">
              <div>{error()}</div>
            </div>
          </Show>
        </div>
        <div class="modal-action">
          <button 
            class="btn btn-error" 
            disabled={!props.project || deleteConfirmName().trim() !== props.project.Name}
            onClick={() => void confirmDelete()}
          >
            Delete Project
          </button>
          <button class="btn" onClick={handleClose}>{t('common.cancel')}</button>
        </div>
      </div>
    </div>
  )
}

export default DeleteConfirmationModal