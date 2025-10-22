import { createSignal } from 'solid-js'
import type { Component } from 'solid-js'
import { useNavigate } from '@solidjs/router'
import { useI18n } from '../i18n'
import type { Project } from '../types/project'
import ProjectList from '../components/project/ProjectList'
import CreateProjectModal from '../components/project/CreateProjectModal.tsx'

import DeleteConfirmationModal from '../components/DeleteConfirmationModal'
import { useQueryClient } from '@tanstack/solid-query'
import { useApiQuery } from '../api/apiHooksW.ts'
import { listProjectsEndpoint } from '../api/endpoints/projects'

const ProjectListPage: Component = () => {
  const { t } = useI18n()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  // UI state
  const [selectedProject, setSelectedProject] = createSignal<Project | null>(null)
  const [showCreateModal, setShowCreateModal] = createSignal(false)

  const [showDeleteModal, setShowDeleteModal] = createSignal(false)
  const [showApplications, setShowApplications] = createSignal(false) // Toggle for enhanced view
  const [error, setError] = createSignal('')
  const [success, setSuccess] = createSignal('')

  // Only load projects list - no builds or deployments
  const projectsQuery = useApiQuery<Project[]>(['projects'], () => listProjectsEndpoint().url)

  // Adapter to keep Resource-like API for existing components
  const projectsData = () => {
    const data = projectsQuery.data as Project[] | undefined
    console.log('Projects data:', data) // Add build print for debugging
    return data
  }
  // Add loading property for compatibility
  Object.defineProperty(projectsData, 'loading', {
    get: () => projectsQuery.isPending,
    enumerable: true,
    configurable: true
  })

  // Action handlers
  function handleSelectProject(project: Project) {
    // Navigate to project detail page using UID
    navigate(`/projects/${project.uid}`)
  }

  function handleDeleteProject(project: Project) {
    setSelectedProject(project)
    setShowDeleteModal(true)
  }





  return (
    <div class="container mx-auto p-6">
      <div class="mb-6 flex items-center justify-between">
        <div>
          <h1 class="text-3xl font-bold text-base-content">{t('projects.title')}</h1>
          <p class="text-base-content/70 mt-2">{t('projects.description')}</p>
        </div>
        <div class="flex gap-2">
          <div class="form-control">
            <label class="label cursor-pointer">
              <span class="label-text mr-2">显示应用</span>
              <input 
                type="checkbox" 
                class="toggle toggle-primary" 
                checked={showApplications()}
                onInput={(e) => setShowApplications(e.currentTarget.checked)}
              />
            </label>
          </div>
          <button class="btn btn-primary" onClick={() => setShowCreateModal(true)}>
            {t('projects.actions.create')}
          </button>
          <button class="btn" onClick={() => void queryClient.invalidateQueries({ queryKey: ['projects'] })}>
            {t('projects.actions.refresh')}
          </button>
        </div>
      </div>

      {error() && (
        <div class="alert alert-error mb-4">
          <span>{error()}</span>
          <button class="btn btn-ghost btn-sm" onClick={() => setError('')}>×</button>
        </div>
      )}
      {success() && (
        <div class="alert alert-success mb-4">
          <span>{success()}</span>
          <button class="btn btn-ghost btn-sm" onClick={() => setSuccess('')}>×</button>
        </div>
      )}

      {/* Projects list */}
      <ProjectList 
        projectsData={projectsData as any}
        onSelectProject={handleSelectProject}
        onDeleteProject={handleDeleteProject}
        showApplications={showApplications()}
      />

      {/* Modals */}
      <CreateProjectModal
        isOpen={showCreateModal()}
        onClose={() => setShowCreateModal(false)}
        onSuccess={setSuccess}
        onError={setError}
        onRefresh={() => void queryClient.invalidateQueries({ queryKey: ['projects'] })}
      />

  

    

      <DeleteConfirmationModal
        isOpen={showDeleteModal()}
        project={selectedProject()}
        onClose={() => setShowDeleteModal(false)}
        onSuccess={(message) => {
          setSuccess(message)
          setSelectedProject(null)
        }}
        onError={setError}
        onRefresh={() => void queryClient.invalidateQueries({ queryKey: ['projects'] })}
      />
    </div>
  )
}

export default ProjectListPage