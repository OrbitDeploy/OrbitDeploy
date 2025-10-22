import { createSignal, createResource, createEffect } from 'solid-js'
import type { Component } from 'solid-js'
import type { Application } from '../../types/project'
import { useApiMutation } from '../../api/apiHooksW.ts'
import { updateApplicationEndpoint } from '../../api/endpoints'

interface EditApplicationModalProps {
  isOpen: boolean
  projectId: number
  application: Application | null
  onClose: () => void
  onSuccess: (message: string) => void
  onError: (message: string) => void
  onRefresh: () => void
}

const EditApplicationModal: Component<EditApplicationModalProps> = (props) => {
  console.log('EditApplicationModal rendering with props:', props)  // Confirm rendering and log props
  console.log('props.application:', props.application)  // Specifically log the application prop
  // Edit application form - prefilled with current data
  const [editApplication, setEditApplication] = createSignal<Partial<Application & { targetPort?: number }>>({
    name: '',
    description: '',
    targetPort: undefined,
    status: 'stopped',
    volumes: [],
    execCommand: '',
    autoUpdatePolicy: '',
    branch: 'main',
  })

  // Reactively update form when application prop changes
  createEffect(() => {
    if (props.application) {
      setEditApplication({
        name: props.application.name,
        description: props.application.description || '',
        targetPort: props.application.targetPort,
        status: props.application.status,
        volumes: props.application.volumes || [],
        execCommand: props.application.execCommand || '',
        autoUpdatePolicy: props.application.autoUpdatePolicy || '',
        branch: props.application.branch || 'main',
      })
    }
  })


  // Fetch project details
  const fetchProject = async () => {
    const response = await fetch(`/api/projects/${props.projectId}`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('access_token')}` },
    })
    if (!response.ok) throw new Error('Failed to fetch project')
    return await response.json()
  }

  const [project] = createResource(() => props.projectId, fetchProject)

  // Function to fetch branches
  const fetchBranches = async () => {
    const repoUrl = project()?.GitRepository
    if (!repoUrl || !repoUrl.includes('github.com')) return []
    const response = await fetch(`/api/projects/${props.projectId}/branches?repoUrl=${encodeURIComponent(repoUrl)}`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('access_token')}` },
    })
    if (!response.ok) throw new Error('Failed to fetch branches')
    const data = await response.json()
    return data.branches || []
  }

  // Resource for branches
  const [branches] = createResource(() => project()?.GitRepository, fetchBranches)

  // Update application mutation
  const updateMutation = useApiMutation<unknown, any>(
    () => props.application?.uid ? updateApplicationEndpoint(props.application.uid) : null,
    {
      onSuccess: () => {
        props.onClose()
        props.onSuccess('应用更新成功')
        props.onRefresh()
      },
      onError: (err: Error) => {
        props.onError(err.message)
      }
    }
  )

  function updateApplication() {
    const form = editApplication()
    const port = form.targetPort ?? 3000
    if (isNaN(port) || port < 1 || port > 65535) {
      props.onError('目标端口必须在1-65535范围内')
      return
    }

    // Create payload matching the API handler format
    const payload = {
      description: form.description?.trim() || '',
      targetPort: form.targetPort,
      status: form.status,
      volumes: form.volumes || {},
      execCommand: form.execCommand?.trim() || null,
      autoUpdatePolicy: form.autoUpdatePolicy?.trim() || null,
      branch: form.branch?.trim() || 'main',
    }

    updateMutation.mutate(payload)
  }

  return (
    <div class={`modal ${props.isOpen ? 'modal-open' : ''}`}>
      <div class="modal-box max-w-2xl">
        <h3 class="font-bold text-lg mb-4">编辑应用</h3>
        <div class="grid grid-cols-1 gap-3">
          <div>
            <div class="flex items-center justify-between mb-2">
              <label class="label">
                <span class="label-text">应用名称 *</span>
              </label>
              <span class="badge badge-ghost badge-sm text-xs">不可修改</span>
            </div>
            <input
              class="input input-bordered input-disabled w-full"
              value={editApplication().name || ''}
              disabled
            />
          </div>

          <div>
            <label class="label"><span class="label-text">应用描述</span></label>
            <textarea
              class="textarea textarea-bordered w-full"
              value={editApplication().description || ''}
              onInput={(e) => setEditApplication(p => ({ ...p, description: e.currentTarget.value }))}
              placeholder="输入应用描述"
              rows="3"
            />
          </div>

          <div>
            <label class="label"><span class="label-text">目标端口 *</span></label>
            <input
              class="input input-bordered w-full"
              type="number"
              min="1"
              max="65535"
              value={editApplication().targetPort ?? ''}
              onInput={(e) => {
                const val = e.currentTarget.value
                setEditApplication(p => ({ ...p, targetPort: val === '' ? undefined : parseInt(val) }))
              }}
              placeholder="容器内部监听的端口号"
            />
          </div>

          <div>
            <label class="label"><span class="label-text">分支</span></label>
            {project.loading ? (
              <div class="flex items-center">
                <span class="loading loading-spinner loading-sm mr-2"></span>
                <span>加载项目信息...</span>
              </div>
            ) : project()?.GitRepository?.includes('github.com') ? (
              <select
                class="select select-bordered w-full"
                value={editApplication().branch || 'main'}
                onInput={(e) => setEditApplication(p => ({ ...p, branch: e.currentTarget.value }))}
                disabled={branches.loading}
              >
                <option value="main">main</option>
                {branches()?.map((branch: string) => <option value={branch}>{branch}</option>)}
              </select>
            ) : (
              <input
                class="input input-bordered w-full"
                value={editApplication().branch || 'main'}
                onInput={(e) => setEditApplication(p => ({ ...p, branch: e.currentTarget.value }))}
                placeholder="输入分支名称"
              />
            )}
          </div>

          <div>
            <label class="label"><span class="label-text">启动命令</span></label>
            <input
              class="input input-bordered w-full"
              value={editApplication().execCommand || ''}
              onInput={(e) => setEditApplication(p => ({ ...p, execCommand: e.currentTarget.value }))}
              placeholder="可选的容器启动命令"
            />
          </div>

          <div>
            <label class="label"><span class="label-text">自动更新策略</span></label>
            <select
              class="select select-bordered w-full"
              value={editApplication().autoUpdatePolicy || ''}
              onInput={(e) => setEditApplication(p => ({ ...p, autoUpdatePolicy: e.currentTarget.value }))}
            >
              <option value="">无自动更新</option>
              <option value="registry">镜像仓库更新</option>
            </select>
          </div>
        </div>

        <div class="modal-action">
          <button
            class="btn btn-primary"
            disabled={updateMutation.isPending}
            onClick={() => void updateApplication()}
          >
            {updateMutation.isPending ? '更新中...' : '更新应用'}
          </button>
          <button class="btn" onClick={props.onClose}>取消</button>
        </div>
      </div>
    </div>
  )
}

export default EditApplicationModal