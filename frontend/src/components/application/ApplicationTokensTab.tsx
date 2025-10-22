import { Component, For, Show, createSignal, createEffect } from 'solid-js'
import type { Application } from '../../types/project'
import { useI18n } from '../../i18n'
import { useApiQuery, useApiMutation } from '../../api/apiHooksW.ts'
import { getApplicationTokensEndpoint, deleteApplicationTokenEndpoint } from '../../api/endpoints'
import CreateApplicationTokenModal from './CreateApplicationTokenModal'

interface ApplicationToken {
  uid : string
  applicationUid: string
  name: string
  expires_at?: string
  last_used_at?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

interface ApplicationTokensTabProps {
  application: Application
}

const ApplicationTokensTab: Component<ApplicationTokensTabProps> = (props) => {
  const { t } = useI18n()
  const [isCreateModalOpen, setIsCreateModalOpen] = createSignal(false)
  const [message, setMessage] = createSignal('')
  const [messageType, setMessageType] = createSignal<'success' | 'error'>('success')

  // Format date utility
  const formatDate = (dateString?: string) => {
    if (!dateString) return '从未使用'
    return new Date(dateString).toLocaleString('zh-CN')
  }

  // Get application tokens
  const tokensQuery = useApiQuery<ApplicationToken[]>(
    () => ['app-tokens', props.application.uid],
    () => getApplicationTokensEndpoint(props.application.uid).url
  )

  // Delete token mutation
  const deleteMutation = useApiMutation<unknown, { appUid: string; tokenUid: string }>(
    (variables: { appUid: string; tokenUid: string }) => 
      deleteApplicationTokenEndpoint(variables.appUid, variables.tokenUid),
    {
      onSuccess: () => {
        showMessage('Token deleted successfully', 'success')
        tokensQuery.refetch()
      },
      onError: (error) => {
        showMessage(error.message, 'error')
      }
    }
  )

  const showMessage = (msg: string, type: 'success' | 'error') => {
    setMessage(msg)
    setMessageType(type)
    setTimeout(() => setMessage(''), 3000)
  }

  const handleDelete = ( tokenUid : string, tokenName: string) => {
    if (window.confirm(`确定要删除 Token "${tokenName}" 吗？删除后将无法恢复。`)) {
      deleteMutation.mutate({ appUid: props.application.uid, tokenUid })
    }
  }

  return (
    <div class="space-y-6">
      {/* Header */}
      <div class="flex justify-between items-center">
        <div>
          <h3 class="text-lg font-semibold">应用 Token 管理</h3>
          <p class="text-sm text-gray-600 mt-1">
            管理用于 CLI 和 API 访问的应用专用 Token
          </p>
        </div>
        <button 
          class="btn btn-primary"
          onClick={() => setIsCreateModalOpen(true)}
        >
          新建 Token
        </button>
      </div>

      {/* Message display */}
      <Show when={message()}>
        <div class={`alert ${messageType() === 'error' ? 'alert-error' : 'alert-success'}`}>
          <span>{message()}</span>
        </div>
      </Show>

      {/* Usage Instructions */}
      <div class="card bg-base-200">
        <div class="card-body p-4">
          <h4 class="font-medium mb-2">使用说明</h4>
          <div class="text-sm space-y-2">
            <p>
              <strong>1. CLI 使用：</strong>
              设置环境变量 ORBIT_TOKEN=your_token 或在命令中使用 --token 参数
            </p>
            <p>
              <strong>2. API 调用：</strong>
              在请求头中添加 Authorization: Bearer your_token
            </p>
            <p>
              <strong>3. 权限范围：</strong>
              Token 仅能访问当前应用 ({props.application.name}) 的资源
            </p>
            <p>
              <strong>4. 安全提示：</strong>
              Token 创建后仅显示一次，请妥善保存。建议设置合理的过期时间。
            </p>
          </div>
        </div>
      </div>

      {/* Token List */}
      <Show when={tokensQuery.isLoading}>
        <div class="flex justify-center py-8">
          <span class="loading loading-spinner loading-md"></span>
        </div>
      </Show>

      <Show when={tokensQuery.error}>
        <div class="alert alert-error">
          <span>加载失败: {tokensQuery.error.message}</span>
        </div>
      </Show>

      <Show when={tokensQuery.data && tokensQuery.data.length === 0}>
        <div class="text-center py-8 text-gray-500">
          <p>暂无 Token</p>
          <p class="text-sm mt-1">点击上方按钮创建第一个 Token</p>
        </div>
      </Show>

      <Show when={tokensQuery.data && tokensQuery.data.length > 0}>
        <div class="space-y-4">
          <For each={tokensQuery.data}>
            {(token) => (
              <div class="card bg-base-100 border">
                <div class="card-body">
                  <div class="flex justify-between items-start">
                    <div class="flex-1">
                      <h4 class="font-semibold text-lg">{token.name}</h4>
                      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4 text-sm">
                        <div>
                          <span class="text-gray-500">状态:</span>
                          <div class="mt-1">
                            <span class={`badge ${token.is_active ? 'badge-success' : 'badge-error'}`}>
                              {token.is_active ? '活跃' : '已禁用'}
                            </span>
                          </div>
                        </div>
                        <div>
                          <span class="text-gray-500">最后使用:</span>
                          <div class="mt-1">{formatDate(token.last_used_at)}</div>
                        </div>
                        <div>
                          <span class="text-gray-500">创建时间:</span>
                          <div class="mt-1">{formatDate(token.created_at)}</div>
                        </div>
                      </div>
                      <Show when={token.expires_at}>
                        <div class="mt-2 text-sm">
                          <span class="text-gray-500">过期时间:</span>
                          <span class="ml-2">{formatDate(token.expires_at)}</span>
                          <Show when={new Date(token.expires_at!) < new Date()}>
                            <span class="badge badge-warning ml-2">已过期</span>
                          </Show>
                        </div>
                      </Show>
                    </div>
                    <div class="flex space-x-2">
                      <button 
                        class="btn btn-error btn-sm"
                        onClick={() => handleDelete(token.uid, token.name)}
                        disabled={deleteMutation.isPending}
                      >
                        删除
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </For>
        </div>
      </Show>

      {/* Create Token Modal */}
      <CreateApplicationTokenModal
        isOpen={isCreateModalOpen()}
        onClose={() => setIsCreateModalOpen(false)}
        applicationUid={props.application.uid}
        applicationName={props.application.name}
        onSuccess={(message) => {
          showMessage(message, 'success')
          tokensQuery.refetch()
        }}
        onError={(message) => showMessage(message, 'error')}
      />
    </div>
  )
}

export default ApplicationTokensTab