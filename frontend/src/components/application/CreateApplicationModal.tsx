import { createSignal, createResource, Show, createEffect } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n } from '../../i18n'
import type { Application } from '../../types/project'
import { useApiMutation } from '../../api/apiHooksW.ts'
import { createAppEndpoint, listProviderAuthsEndpoint, getProviderAuthRepositoriesEndpoint } from '../../api/endpoints'

interface CreateApplicationModalProps {
  isOpen: boolean
  projectUid: string
  projectName: string
  onClose: () => void
  onSuccess: (message: string) => void
  onError: (message: string) => void
  onRefresh: () => void
}

const CreateApplicationModal: Component<CreateApplicationModalProps> = (props) => {
  const { t } = useI18n()

  // Create application form
  const [newApplication, setNewApplication] = createSignal<Partial<Application & { targetPort?: number; repoUrl?: string; buildDir?: string; buildType?: string; providerAuthUid?: string; repoFullName?: string }>>({
    name: '',
    description: '',
    targetPort: 3000, 
    volumes: [],
    execCommand: '',
    autoUpdatePolicy: '',
    branch: 'main',
    repoUrl: '',
    buildDir: '/',
    buildType: 'dockerfile',
    providerAuthUid: undefined,
    repoFullName: '',
  })

  // Fetch project details
  const fetchProject = async () => {
    const response = await fetch(`/api/projects/${props.projectUid}`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('access_token')}` },
    })
    if (!response.ok) throw new Error('Failed to fetch project')
    return await response.json()
  }

  const [project] = createResource(() => props.projectUid, fetchProject)

  // Fetch provider auths
  const fetchProviderAuths = async () => {
    const response = await fetch(listProviderAuthsEndpoint().url, {
      headers: { Authorization: `Bearer ${localStorage.getItem('access_token')}` },
    })
    if (!response.ok) throw new Error('Failed to fetch provider auths')
    const data = await response.json()
    return data.data || []
  }

  const [providerAuths] = createResource(fetchProviderAuths)

  // Function to fetch branches
  const fetchBranches = async ({ providerAuthUid, repo }: { providerAuthUid?: string; repo?: string }) => {
    if (!providerAuthUid || !repo) return []
    const response = await fetch(`/api/provider-auths/${providerAuthUid}/repositories/branches?repo=${encodeURIComponent(repo)}`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('access_token')}` },
    })
    if (!response.ok) throw new Error('Failed to fetch branches')
    const data = await response.json()
    return data.data || []
  }

  // Resource for branches
  const [branches] = createResource(() => ({ providerAuthUid: newApplication().providerAuthUid, repo: newApplication().repoFullName }), fetchBranches)

  // Fetch repositories for selected provider auth
  const fetchRepositories = async (providerAuthUid: string) => {
    const response = await fetch(getProviderAuthRepositoriesEndpoint(providerAuthUid).url, {
      headers: { Authorization: `Bearer ${localStorage.getItem('access_token')}` },
    })
    if (!response.ok) throw new Error('Failed to fetch repositories')
    const data = await response.json()
    return data.data || []
  }

  const [repositories] = createResource(() => newApplication().providerAuthUid, (uid) => uid ? fetchRepositories(uid) : Promise.resolve([]))

  // Helper to get base URL for platform
  const getBaseUrl = (platform: string) => {
    switch (platform) {
      case 'github':
        return 'https://github.com/'
      case 'gitlab':
        return 'https://gitlab.com/'
      case 'bitbucket':
        return 'https://bitbucket.org/'
      case 'gitea':
        // For gitea, assume full URL is provided or use redirectURI, but for now, require manual input
        return ''
      default:
        return ''
    }
  }

  // Create application mutation
  const createMutation = useApiMutation<unknown, { payload: any }>(
    () => createAppEndpoint(props.projectUid),
    {
      onSuccess: () => {
        props.onClose()
        resetForm()
        props.onSuccess('应用创建成功')
        props.onRefresh()
      },
      onError: (err: Error) => {
        setErrorMessage(err.message)
      }
    }
  )

  const [errorMessage, setErrorMessage] = createSignal('')

  createEffect(() => {
    if (props.isOpen) {
      setErrorMessage('')
      resetForm()  // Reset the form when the modal opens
    }
  })

  // Function to generate a random 5-letter string
  const generateRandomString = (length: number) => {
    const chars = 'abcdefghijklmnopqrstuvwxyz'
    let result = ''
    for (let i = 0; i < length; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length))
    }
    return result
  }

  function createApplication() {
    const form = newApplication()
    setErrorMessage('') // Clear previous error
    if (!form.name?.trim()) {
      setErrorMessage('应用名称是必填项')
      return
    }

    const port = form.targetPort ?? 3000
    if (isNaN(port) || port < 1 || port > 65535) {
      setErrorMessage('目标端口必须在1-65535范围内')
      return
    }
    
    // Create payload matching the API handler format
    const payload = {
      name: form.name.trim(),
      description: form.description?.trim() || '',
      repoUrl: form.repoUrl?.trim() || null,
      targetPort: port,
      volumes: form.volumes || {},
      execCommand: form.execCommand?.trim() || null,
      autoUpdatePolicy: form.autoUpdatePolicy?.trim() || null,
      branch: form.branch?.trim() || 'main',
      buildDir: form.buildDir?.trim() || '/',
      buildType: form.buildType?.trim() || 'dockerfile',
      providerAuthUid: form.providerAuthUid || null,
    }

    createMutation.mutate(payload)
  }

  const resetForm = () => {
    const randomSuffix = generateRandomString(5)
    const projName = (props.projectName?.toLowerCase() || '') + "-" + randomSuffix
    setNewApplication({
      name: projName,
      description: '',
      repoUrl: '',
      targetPort: 3000, 
      volumes: [],
      execCommand: '',
      autoUpdatePolicy: '',
      branch: 'main',
      buildDir: '/',
      buildType: 'dockerfile',
      providerAuthUid: undefined,
      repoFullName: '',
    })
  }

  return (
    <div class={`modal ${props.isOpen ? 'modal-open' : ''}`}>
      <div class="modal-box max-w-2xl">
        <h3 class="font-bold text-lg mb-4">新建应用</h3>
        <div class="grid grid-cols-1 gap-3">
          <div>
            <label class="label">
              <span class="label-text">应用名称 *</span>
              <span class="label-text-alt text-sm">只支持小写字母和-连接符，系统内唯一。</span>
            </label>
            <input 
              class="input input-bordered w-full" 
              value={newApplication().name || ''} 
              onInput={(e) => setNewApplication(p => ({ ...p, name: e.currentTarget.value }))}
              placeholder="输入应用名称" 
            />
          </div>
          
          <div>
            <label class="label"><span class="label-text">应用描述</span></label>
            <textarea 
              class="textarea textarea-bordered w-full" 
              value={newApplication().description || ''} 
              onInput={(e) => setNewApplication(p => ({ ...p, description: e.currentTarget.value }))}
              placeholder="输入应用描述"
              rows="3"
            />
          </div>
          
          <div>
            <label class="label"><span class="label-text">仓库授权</span></label>
            <select 
              class="select select-bordered w-full"
              value={newApplication().providerAuthUid || ''}
              onInput={(e) => {
                const val = e.currentTarget.value
                setNewApplication(p => ({ ...p, providerAuthUid: val ? val : undefined, repoUrl: '' }))
              }}
            >
              <option value="">无（支持CLI推送）</option>
              {providerAuths()?.map((auth: any) => <option value={auth.uid}>{auth.platform} - {auth.appId || auth.clientId || auth.username}</option>)}
            </select>
          </div>
          
          <div>
            <label class="label"><span class="label-text">仓库URL</span></label>
            {newApplication().providerAuthUid ? (
              <select 
                class="select select-bordered w-full"
                value={newApplication().repoFullName || ''}
                onInput={(e) => {
                  const selectedRepo = e.currentTarget.value
                  const providerAuth = providerAuths()?.find((auth: any) => auth.uid === newApplication().providerAuthUid)
                  const baseUrl = providerAuth ? getBaseUrl(providerAuth.platform) : ''
                  const fullUrl = baseUrl ? baseUrl + selectedRepo : selectedRepo
                  setNewApplication(p => ({ ...p, repoUrl: fullUrl, repoFullName: selectedRepo }))
                }}
                disabled={repositories.loading}
              >
                <option value="">选择仓库</option>
                {repositories()?.map((repo: any) => <option value={repo.fullName}>{repo.fullName}</option>)}
              </select>
            ) : (
              <input 
                class="input input-bordered w-full" 
                onInput={(e) => setNewApplication(p => ({ ...p, repoUrl: e.currentTarget.value, repoFullName: '' }))
                }
                placeholder="https://github.com/user/repo"
              />
            )}
          </div>
          
          <div>
            <label class="label"><span class="label-text">分支</span></label>
            {newApplication().providerAuthUid && newApplication().repoUrl ? (
              <select 
                class="select select-bordered w-full"
                value={newApplication().branch || ''}
                onInput={(e) => setNewApplication(p => ({ ...p, branch: e.currentTarget.value }))}
                disabled={branches.loading}
              >
                <option value="">选择分支</option>
                {branches()?.map((branch: any) => <option value={branch.name}>{branch.name}</option>)}
              </select>
            ) : (
              <input 
                class="input input-bordered w-full" 
                value={newApplication().branch || ''} 
                onInput={(e) => setNewApplication(p => ({ ...p, branch: e.currentTarget.value }))}
                placeholder="输入分支名称"
              />
            )}
          </div>
          
          <div>
            <label class="label"><span class="label-text">构建目录</span></label>
            <input 
              class="input input-bordered w-full" 
              value={newApplication().buildDir || '/'} 
              onInput={(e) => setNewApplication(p => ({ ...p, buildDir: e.currentTarget.value }))}
              placeholder="构建目录，默认根目录"
            />
          </div>
          
          <div>
            <label class="label"><span class="label-text">构建类型</span></label>
            <select 
              class="select select-bordered w-full"
              value={newApplication().buildType || 'dockerfile'}
              onInput={(e) => setNewApplication(p => ({ ...p, buildType: e.currentTarget.value }))}
            >
              <option value="dockerfile">Dockerfile</option>
              {/* <option value="railpack">Railpack</option>
              <option value="nixpacks">Nixpacks</option> */}
            </select>
          </div>
          
          <div>
            <label class="label"><span class="label-text">目标端口 *</span></label>
            <input 
              class="input input-bordered w-full" 
              type="number"
              min="1"
              max="65535"
              value={newApplication().targetPort ?? ''} 
              onInput={(e) => {
                const val = e.currentTarget.value
                setNewApplication(p => ({ ...p, targetPort: val === '' ? undefined : parseInt(val) }))
              }}
              placeholder="容器内部监听的端口号"
            />
          </div>
          
          <div>
            <label class="label"><span class="label-text">启动命令</span></label>
            <input 
              class="input input-bordered w-full" 
              value={newApplication().execCommand || ''} 
              onInput={(e) => setNewApplication(p => ({ ...p, execCommand: e.currentTarget.value }))}
              placeholder="可选的容器启动命令"
            />
          </div>
          
          <div>
            <label class="label"><span class="label-text">自动更新策略</span></label>
            <select 
              class="select select-bordered w-full"
              value={newApplication().autoUpdatePolicy || ''}
              onInput={(e) => setNewApplication(p => ({ ...p, autoUpdatePolicy: e.currentTarget.value }))}
            >
              <option value="">无自动更新</option>
              <option value="registry">镜像仓库更新</option>
            </select>
          </div>
        </div>
        
        <Show when={errorMessage()}>
          <div class="alert alert-error mb-4">
            <span>{errorMessage()}</span>
          </div>
        </Show>
        
        <div class="modal-action">
          <button 
            class="btn btn-primary" 
            disabled={createMutation.isPending} 
            onClick={() => void createApplication()}
          >
            {createMutation.isPending ? '创建中...' : '创建应用'}
          </button>
          <button class="btn" onClick={props.onClose}>取消</button>
        </div>
      </div>
    </div>
  )
}

export default CreateApplicationModal