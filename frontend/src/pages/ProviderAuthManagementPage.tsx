import { createSignal, createMemo, onMount, For, Show, createEffect } from 'solid-js'
import type { Component } from 'solid-js'
import { useI18n } from '../i18n'
import type { ProviderAuth, CreateProviderAuthRequest } from '../types/providerAuth'
import { useApiMutation, useApiQuery } from '../api/apiHooksW.ts'
import { 
  listProviderAuthsEndpoint, 
  createProviderAuthEndpoint, 
  updateProviderAuthEndpoint, 
  deleteProviderAuthEndpoint,
  activateProviderAuthEndpoint,
  deactivateProviderAuthEndpoint
} from '../api/endpoints/providerAuths'
import GitHubAppIntegrationModal from '../components/GitHubAppIntegrationModal'
import CreateEditProviderAuthModal from '../components/CreateEditProviderAuthModal'



const ProviderAuthManagementPage: Component = () => {
  const { t } = useI18n()
  const [isCreateModalOpen, setIsCreateModalOpen] = createSignal(false)
  const [isGitHubAppModalOpen, setIsGitHubAppModalOpen] = createSignal(false)
  const [editingAuth, setEditingAuth] = createSignal<ProviderAuth | null>(null)
  const [message, setMessage] = createSignal('')
  const [messageType, setMessageType] = createSignal<'success' | 'error'>('success')
  const [selectedPlatform, setSelectedPlatform] = createSignal<'github' | 'gitlab' | 'bitbucket' | 'gitea'>('github')

  // 表单状态
  const [formData, setFormData] = createSignal<Partial<CreateProviderAuthRequest>>({
    platform: 'github',
    clientId: '',
    clientSecret: '',
    redirectUri: '',
    username: '',
    appPassword: '',
    scopes: '',
    isActive: true
  })

  // ==================== [代码修复] ====================
  // 1. 不要解构 useApiQuery 的返回值，而是保留整个 query 对象
  const authsQuery = useApiQuery<ProviderAuth[]>(
    ['provider-auths'],
    () => listProviderAuthsEndpoint().url
  )

  // 2. 创建响应式的访问器函数来读取状态
  const auths = () => authsQuery.data || []
  // 使用 isPending 表示初次加载，这在 TanStack Query v5 中是最佳实践
  const isLoading = () => authsQuery.isPending
  const error = () => authsQuery.error
  const refetchAuths = () => authsQuery.refetch()
  // ======================================================



  const showMessage = (msg: string, type: 'success' | 'error') => {
    setMessage(msg)
    setMessageType(type)
    setTimeout(() => setMessage(''), 5000)
  }

  // 创建授权
  const createMutation = useApiMutation<ProviderAuth, CreateProviderAuthRequest>(
    createProviderAuthEndpoint(),
    {
      onSuccess: () => {
        showMessage(t('providerAuth.createSuccess'), 'success')
        setIsCreateModalOpen(false)
        resetForm()
        void refetchAuths() // 使用 void 明确表示我们不关心 promise 的结果
      },
      onError: (err) => {
        showMessage(t('providerAuth.createFailed', { message: err.message }), 'error')
      }
    }
  )

  // 更新授权
  const updateMutation = useApiMutation<ProviderAuth, Partial<CreateProviderAuthRequest> & { id: string }>(
    ({ id, ...data }) => updateProviderAuthEndpoint(id),
    {
      body: ({ id, ...data }) => data,
      onSuccess: () => {
        showMessage(t('providerAuth.updateSuccess'), 'success')
        setEditingAuth(null)
        resetForm()
        void refetchAuths()
      },
      onError: (err) => {
        showMessage(t('providerAuth.updateFailed', { message: err.message }), 'error')
      }
    }
  )

  // 删除授权
  const deleteMutation = useApiMutation<unknown, { id: string }>(
    ({ id }) => deleteProviderAuthEndpoint(id),
    {
      onSuccess: () => {
        showMessage(t('providerAuth.deleteSuccess'), 'success')
        void refetchAuths()
      },
      onError: (err) => {
        showMessage(t('providerAuth.deleteFailed', { message: err.message }), 'error')
      }
    }
  )

  // 激活/停用授权
  const toggleActiveMutation = useApiMutation<unknown, { id: string; action: 'activate' | 'deactivate' }>(
    ({ id, action }) => action === 'activate' ? activateProviderAuthEndpoint(id) : deactivateProviderAuthEndpoint(id),
    {
      onSuccess: (_, variables) => {
        showMessage(t(variables.action === 'activate' ? 'providerAuth.activateSuccess' : 'providerAuth.deactivateSuccess'), 'success')
        void refetchAuths()
      },
      onError: (err) => {
        showMessage(t('providerAuth.operationFailed', { message: err.message }), 'error')
      }
    }
  )

  const resetForm = () => {
    setFormData({
      platform: 'github',
      // repoUrl: '', // 这个字段在您的类型中不存在，可能需要检查一下
      clientId: '',
      clientSecret: '',
      redirectUri: '',
      username: '',
      appPassword: '',
      scopes: '',
      isActive: true
    })
    setSelectedPlatform('github')
  }

  const handleSubmit = () => {
    const data = formData()
    if (editingAuth()) {
      updateMutation.mutate({ ...data, id: editingAuth()!.uid } as any)
    } else {
      createMutation.mutate(data as CreateProviderAuthRequest)
    }
  }

  const startEdit = (auth: ProviderAuth) => {
    setEditingAuth(auth)
    setFormData({
      platform: auth.platform as 'github' | 'gitlab' | 'bitbucket' | 'gitea',
      clientId: auth.clientId,
      clientSecret: '', // 不填充敏感信息
      redirectUri: auth.redirectUri,
      username: auth.username || '', // 处理 null
      appPassword: '', // 不填充敏感信息
      scopes: auth.scopes || '', // 处理 null
      isActive: auth.isActive
    })
    setSelectedPlatform(auth.platform as 'github' | 'gitlab' | 'bitbucket' | 'gitea')
    setIsCreateModalOpen(true)
  }

  const cancelEdit = () => {
    setEditingAuth(null)
    setIsCreateModalOpen(false)
    resetForm()
  }

  const getPlatformBadgeClass = (platform: string) => {
    const classes: Record<string, string> = {
      github: 'badge-primary',
      gitlab: 'badge-warning',
      bitbucket: 'badge-info',
      gitea: 'badge-success'
    }
    return classes[platform] || 'badge-neutral'
  }

  // 新增：处理GitHub App安装跳转
  const handleInstallGitHubApp = (auth: ProviderAuth) => {
    if (auth.platform === 'github' && auth.slug) {
      const installUrl = `https://github.com/apps/${auth.slug}/installations/new`
      window.open(installUrl, '_blank')
    } else {
      showMessage(t('providerAuth.cannotGenerateInstallLink'), 'error')
    }
  }

  onMount(() => {
    // Check if we're returning from GitHub App creation
    const urlParams = new URLSearchParams(window.location.search)
    if (urlParams.get('github_app_created') === 'true') {
      showMessage(t('providerAuth.githubAppIntegrationMessage'), 'success')
      setIsGitHubAppModalOpen(true)
      // Clean up the URL
      window.history.replaceState({}, document.title, window.location.pathname)
    }
  })

  return (
    <div class="container mx-auto p-6">
      <div class="flex justify-between items-center mb-6">
        <div>
          <h1 class="text-3xl font-bold mb-2">{t('providerAuth.pageTitle')}</h1>
          <p class="text-base-content/60">
            {t('providerAuth.pageDescription')}
          </p>
        </div>
        <div class="flex gap-2">
          <button
            class="btn btn-primary"
            onClick={() => {
              resetForm()
              setEditingAuth(null) // 确保点击添加时清除编辑状态
              setIsCreateModalOpen(true)
            }}
          >
            <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            {t('providerAuth.addAuth')}
          </button>
          <button
            class="btn btn-secondary"
            onClick={() => setIsGitHubAppModalOpen(true)}
          >
            <svg class="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
            </svg>
            {t('providerAuth.githubAppIntegration')}
          </button>
        </div>
      </div>

      {/* 消息提示 */}
      <Show when={message()}>
        <div class={`alert ${messageType() === 'success' ? 'alert-success' : 'alert-error'} mb-4`}>
          <span>{message()}</span>
        </div>
      </Show>

      {/* 授权列表 */}
      <div class="card bg-base-100 shadow-xl">
        <div class="card-body">
          <h2 class="card-title mb-4">{t('providerAuth.authList')}</h2>
          
          {/* 在 JSX 中，现在需要使用 isLoading() 和 error() */}
          <Show 
            when={!isLoading() && auths().length > 0}
            fallback={
              <div class="text-center py-8">
                <Show when={isLoading()}>
                  <div class="loading loading-spinner loading-lg"></div>
                  <p class="mt-2">{t('providerAuth.loading')}</p>
                </Show>
                <Show when={!isLoading() && !error() && auths().length === 0}>
                  <div class="text-base-content/60">
                    <svg class="w-16 h-16 mx-auto mb-4 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <p class="text-lg mb-2">{t('providerAuth.noAuth')}</p>
                    <p>{t('providerAuth.noAuthDescription')}</p>
                  </div>
                </Show>
                <Show when={error()}>
                  <div class="alert alert-error">
                    <span>{t('providerAuth.loadFailed', { message: error()?.message })}</span>
                  </div>
                </Show>
              </div>
            }
          >
            <div class="overflow-x-auto">
              <table class="table table-zebra w-full">
                <thead>
                  <tr>
                    <th>{t('providerAuth.platform')}</th>
                    <th>{t('providerAuth.clientIdOrUsername')}</th>
                    <th>{t('providerAuth.status')}</th>
                    <th>{t('providerAuth.creationTime')}</th>
                    <th>{t('providerAuth.actions')}</th>
                  </tr>
                </thead>
                <tbody>
                  <For each={auths()}>
                    {(auth) => (
                      <tr>
                        <td>
                          <div class={`badge ${getPlatformBadgeClass(auth.platform)}`}>
                            {t(`platform.${auth.platform}`)}
                          </div>
                        </td>
                      
                        <td>
                          <div class="max-w-xs truncate" title={auth.clientId || auth.username || 'N/A'}>
                            {auth.clientId || auth.username || 'N/A'}
                          </div>
                        </td>
                        <td>
                          <div class={`badge ${auth.isActive ? 'badge-success' : 'badge-error'}`}>
                            {auth.isActive ? t('providerAuth.enabled') : t('providerAuth.disabled')}
                          </div>
                        </td>
                        <td>
                          {new Date(auth.createdAt).toLocaleDateString()}
                        </td>
                        <td>
                          <div class="flex gap-2">
                            <button
                              class="btn btn-sm btn-ghost"
                              onClick={() => startEdit(auth)}
                            >
                              {t('providerAuth.edit')}
                            </button>
                            {/* 新增：GitHub Apps安装按钮 */}
                            <Show when={auth.platform === 'github' && auth.slug}>
                              <button
                                class="btn btn-sm btn-primary"
                                onClick={() => handleInstallGitHubApp(auth)}
                              >
                                {t('providerAuth.startInstallation')}
                              </button>
                            </Show>
                            <button
                              class={`btn btn-sm ${auth.isActive ? 'btn-warning' : 'btn-success'}`}
                              onClick={() => toggleActiveMutation.mutate({ 
                                id: auth.uid, 
                                action: auth.isActive ? 'deactivate' : 'activate' 
                              })}
                            >
                              {auth.isActive ? t('providerAuth.disable') : t('providerAuth.enable')}
                            </button>
                            <button
                              class="btn btn-sm btn-error"
                              onClick={() => {
                                if (confirm(t('providerAuth.deleteConfirm'))) {
                                  deleteMutation.mutate({ id: auth.uid })
                                }
                              }}
                            >
                              {t('providerAuth.delete')}
                            </button>
                          </div>
                        </td>
                      </tr>
                    )}
                  </For>
                </tbody>
              </table>
            </div>
          </Show>
        </div>
      </div>

      {/* 创建/编辑模态框 */}
      <CreateEditProviderAuthModal
        isOpen={isCreateModalOpen()}
        onClose={cancelEdit} // 使用 cancelEdit 保证状态重置
        editingAuth={editingAuth()}
        formData={formData()}
        setFormData={setFormData}
        selectedPlatform={selectedPlatform()}
        setSelectedPlatform={setSelectedPlatform}
        handleSubmit={handleSubmit}
        cancelEdit={cancelEdit}
        createMutation={createMutation}
        updateMutation={updateMutation}
      />

      {/* GitHub Apps Integration Modal */}
      <GitHubAppIntegrationModal
        isOpen={isGitHubAppModalOpen()}
        onClose={() => setIsGitHubAppModalOpen(false)}
        onSuccess={(message) => {
          showMessage(message, 'success')
          setIsGitHubAppModalOpen(false)
          void refetchAuths()
        }}
        onError={(message) => showMessage(message, 'error')}
        userId={1} // TODO: Get from auth context
      />
    </div>
  )
}

export default ProviderAuthManagementPage